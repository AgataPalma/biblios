#!/usr/bin/env node

import fs from 'node:fs';
import path from 'node:path';
import process from 'node:process';
import { fileURLToPath } from 'node:url';

const scriptDirectory = path.dirname(fileURLToPath(import.meta.url));
const repositoryRoot = path.resolve(scriptDirectory, '..');

const files = {
  router: path.join(repositoryRoot, 'backend/cmd/api/main.go'),
  openapi: path.join(repositoryRoot, 'backend/cmd/api/docs/openapi.yaml'),
  collection: path.join(repositoryRoot, 'biblios-api.postman_collection.json'),
  localEnvironment: path.join(repositoryRoot, 'api-tests/postman/environments/local.postman_environment.json'),
  baseline: path.join(repositoryRoot, 'api-tests/postman/openapi-gap-baseline.json'),
};

const supportedMethods = new Set(['GET', 'POST', 'PUT', 'PATCH', 'DELETE']);
const argumentsSet = new Set(process.argv.slice(2));
const strictOpenApi = argumentsSet.has('--strict-openapi');
const printBaseline = argumentsSet.has('--print-openapi-baseline');

function readText(filename) {
  return fs.readFileSync(filename, 'utf8');
}

function readJson(filename) {
  return JSON.parse(readText(filename));
}

function normalizePathname(value) {
  let pathname = value.split('?', 1)[0].trim();
  if (!pathname.startsWith('/')) pathname = `/${pathname}`;
  pathname = pathname.replace(/\/+/g, '/');
  return pathname.length > 1 ? pathname.replace(/\/$/, '') : pathname;
}

function canonicalPath(value) {
  return normalizePathname(value)
    .split('/')
    .map((segment) => (/^(\{[^}]+\}|\{\{[^}]+\}\}|:[^/]+)$/.test(segment) ? '{}' : segment))
    .join('/');
}

function operationKey(method, pathname) {
  return `${method.toUpperCase()} ${canonicalPath(pathname)}`;
}

function lineNumber(source, index) {
  return source.slice(0, index).split('\n').length;
}

function extractRouterOperations(source) {
  const operations = [];
  const healthMatch = /\br\.Get\("\/health"/.exec(source);
  if (!healthMatch) throw new Error('Unable to find the top-level GET /health route.');
  operations.push({ method: 'GET', path: '/health', line: lineNumber(source, healthMatch.index) });

  const registerStart = source.indexOf('func registerRoutes(');
  if (registerStart === -1) throw new Error('Unable to find func registerRoutes.');
  const registerSource = source.slice(registerStart);
  const routePattern = /\b(Get|Post|Put|Patch|Delete)\("([^"]+)"/g;
  for (const match of registerSource.matchAll(routePattern)) {
    operations.push({
      method: match[1].toUpperCase(),
      path: `/api/v1${normalizePathname(match[2])}`,
      line: lineNumber(source, registerStart + match.index),
    });
  }

  return operations;
}

function extractOpenApiOperations(source) {
  const operations = [];
  let currentPath;
  const lines = source.split('\n');

  for (let index = 0; index < lines.length; index += 1) {
    const pathMatch = /^  (\/[^:]+):\s*$/.exec(lines[index]);
    if (pathMatch) {
      currentPath = pathMatch[1];
      continue;
    }

    const methodMatch = /^    (get|post|put|patch|delete):\s*$/.exec(lines[index]);
    if (!methodMatch || !currentPath) continue;
    const fullPath = currentPath === '/health' ? currentPath : `/api/v1${currentPath}`;
    operations.push({ method: methodMatch[1].toUpperCase(), path: fullPath, line: index + 1 });
  }

  return operations;
}

function walkCollection(items, parents = []) {
  const requests = [];
  for (const item of items ?? []) {
    const itemPath = [...parents, item.name ?? '<unnamed>'];
    if (item.request) requests.push({ item, request: item.request, itemPath });
    requests.push(...walkCollection(item.item, itemPath));
  }
  return requests;
}

function collectionVariableMap(collection, environment) {
  const variables = new Map();
  for (const variable of collection.variable ?? []) variables.set(variable.key, String(variable.value ?? ''));
  for (const variable of environment.values ?? []) {
    if (variable.enabled !== false) variables.set(variable.key, String(variable.value ?? ''));
  }
  return variables;
}

function requestUrl(request, variables) {
  const url = typeof request.url === 'string' ? request.url : request.url?.raw;
  if (!url) throw new Error('Request has no URL.');

  const resolved = url.replace(/\{\{([^}]+)\}\}/g, (whole, name) => {
    if (name === 'baseUrl' || name === 'healthUrl') return variables.get(name) || whole;
    return whole;
  });

  if (/^https?:\/\//.test(resolved)) return new URL(resolved).pathname;
  return resolved.split('?', 1)[0];
}

function routerPathMatches(routerPath, requestPath) {
  const pattern = `^${routerPath
    .split('/')
    .map((segment) => (/^\{[^}]+\}$/.test(segment) ? '[^/]+' : segment.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')))
    .join('/')}$`;
  return new RegExp(pattern).test(normalizePathname(requestPath));
}

function testScript(item) {
  return (item.event ?? [])
    .filter((event) => event.listen === 'test')
    .flatMap((event) => event.script?.exec ?? [])
    .join('\n');
}

function collectVariableReferences(value, references = new Set()) {
  if (typeof value === 'string') {
    for (const match of value.matchAll(/\{\{([^}]+)\}\}/g)) references.add(match[1]);
  } else if (Array.isArray(value)) {
    for (const child of value) collectVariableReferences(child, references);
  } else if (value && typeof value === 'object') {
    for (const child of Object.values(value)) collectVariableReferences(child, references);
  }
  return references;
}

function sortedUnique(values) {
  return [...new Set(values)].sort();
}

function compareExactSet(label, actual, expected, errors) {
  const missing = expected.filter((value) => !actual.includes(value));
  const unexpected = actual.filter((value) => !expected.includes(value));
  if (missing.length || unexpected.length) {
    errors.push(`${label} changed. Missing baseline entries: ${missing.join(', ') || 'none'}. New entries: ${unexpected.join(', ') || 'none'}.`);
  }
}

const errors = [];
const warnings = [];

const routerSource = readText(files.router);
const openApiSource = readText(files.openapi);
const collection = readJson(files.collection);
const localEnvironment = readJson(files.localEnvironment);

const routerOperations = extractRouterOperations(routerSource);
const openApiOperations = extractOpenApiOperations(openApiSource);
const requests = walkCollection(collection.item);
const variables = collectionVariableMap(collection, localEnvironment);

const routerKeys = sortedUnique(routerOperations.map(({ method, path: pathname }) => operationKey(method, pathname)));
const openApiKeys = sortedUnique(openApiOperations.map(({ method, path: pathname }) => operationKey(method, pathname)));

if (routerKeys.length !== routerOperations.length) errors.push('The Go router contains duplicate method/path registrations.');
if (openApiKeys.length !== openApiOperations.length) errors.push('OpenAPI contains duplicate method/path operations.');

const missingFromOpenApi = routerKeys.filter((key) => !openApiKeys.includes(key));
const absentFromRouter = openApiKeys.filter((key) => !routerKeys.includes(key));
if (absentFromRouter.length) errors.push(`OpenAPI operations not registered by the router: ${absentFromRouter.join(', ')}`);

if (printBaseline) {
  process.stdout.write(`${JSON.stringify({
    description: 'Known router operations not yet represented in OpenAPI. Reduce this list as BLIO-43 reconciles the contract; CI rejects any unreviewed change.',
    missingFromOpenApi,
  }, null, 2)}\n`);
  process.exit(0);
}

if (strictOpenApi && missingFromOpenApi.length) {
  errors.push(`OpenAPI is missing ${missingFromOpenApi.length} registered operation(s): ${missingFromOpenApi.join(', ')}`);
} else if (fs.existsSync(files.baseline)) {
  const baseline = readJson(files.baseline);
  compareExactSet('The known OpenAPI gap baseline', missingFromOpenApi, baseline.missingFromOpenApi ?? [], errors);
  if (missingFromOpenApi.length) warnings.push(`OpenAPI still has ${missingFromOpenApi.length} baselined gap(s); BLIO-43 remains incomplete until the baseline is empty.`);
} else if (missingFromOpenApi.length) {
  errors.push(`OpenAPI is missing ${missingFromOpenApi.length} operation(s) and no reviewed gap baseline exists.`);
}

const coveredRouterKeys = new Set();
for (const { item, request, itemPath } of requests) {
  const label = itemPath.join(' / ');
  const method = String(request.method ?? '').toUpperCase();
  if (!supportedMethods.has(method)) {
    errors.push(`${label}: unsupported or missing HTTP method ${method || '<empty>'}.`);
    continue;
  }

  let pathname;
  try {
    pathname = requestUrl(request, variables);
  } catch (error) {
    errors.push(`${label}: ${error.message}`);
    continue;
  }

  const matchingOperation = routerOperations.find((operation) => (
    operation.method === method && routerPathMatches(operation.path, pathname)
  ));
  if (!matchingOperation) {
    errors.push(`${label}: ${method} ${pathname} does not match a registered Go route.`);
  } else {
    coveredRouterKeys.add(operationKey(matchingOperation.method, matchingOperation.path));
  }

  const script = testScript(item);
  if (!script.trim()) errors.push(`${label}: no active post-response test script.`);
  if (!/pm\.response\.to\.have\.status\(\d+\)|pm\.expect\(pm\.response\.code\)/.test(script)) {
    errors.push(`${label}: no explicit HTTP status assertion.`);
  }

  for (const header of request.header ?? []) {
    const headerName = String(header.key ?? '').toLowerCase();
    const headerValue = String(header.value ?? '');
    if (['authorization', 'x-api-key', 'api-key'].includes(headerName) && headerValue && !headerValue.includes('{{')) {
      errors.push(`${label}: ${header.key} contains a literal credential-like value.`);
    }
  }
}

const missingFromCollection = routerKeys.filter((key) => !coveredRouterKeys.has(key));
if (missingFromCollection.length) {
  errors.push(`Postman has no request for ${missingFromCollection.length} registered operation(s): ${missingFromCollection.join(', ')}`);
}

const declaredVariables = new Set([
  ...(collection.variable ?? []).map((variable) => variable.key),
  ...(localEnvironment.values ?? []).map((variable) => variable.key),
]);
const runtimeVariables = new Set();
for (const { item } of requests) {
  for (const match of testScript(item).matchAll(/pm\.(?:collectionVariables|environment)\.set\(['"]([^'"]+)['"]/g)) {
    runtimeVariables.add(match[1]);
  }
}
const referencedVariables = collectVariableReferences(collection.item);
const undeclaredVariables = sortedUnique([...referencedVariables].filter((name) => !declaredVariables.has(name) && !runtimeVariables.has(name)));
if (undeclaredVariables.length) errors.push(`Postman references undeclared variables: ${undeclaredVariables.join(', ')}`);

const collectionSecretKeys = /token|password|secret|api.?key/i;
for (const variable of collection.variable ?? []) {
  if (collectionSecretKeys.test(variable.key) && String(variable.value ?? '').trim()) {
    errors.push(`Collection variable ${variable.key} must not contain a committed value.`);
  }
}

const scriptsWithBodyAssertions = requests.filter(({ item }) => /pm\.response\.(?:json|text)\(|pm\.expect\(/.test(testScript(item))).length;
if (scriptsWithBodyAssertions < requests.length) {
  warnings.push(`${requests.length - scriptsWithBodyAssertions} request(s) assert status only; BLIO-47 must classify which require schema or semantic assertions.`);
}

console.log('Biblios API contract validation');
console.log(`  Go router operations:       ${routerKeys.length}`);
console.log(`  OpenAPI operations:         ${openApiKeys.length}`);
console.log(`  Postman requests:           ${requests.length}`);
console.log(`  Router operations covered:  ${coveredRouterKeys.size}/${routerKeys.length}`);
console.log(`  Requests with test scripts: ${requests.filter(({ item }) => testScript(item).trim()).length}/${requests.length}`);
console.log(`  Requests with body checks:  ${scriptsWithBodyAssertions}/${requests.length}`);

for (const warning of warnings) console.warn(`WARNING: ${warning}`);
for (const error of errors) console.error(`ERROR: ${error}`);

if (errors.length) {
  console.error(`Validation failed with ${errors.length} error(s).`);
  process.exit(1);
}

console.log('Validation passed.');
