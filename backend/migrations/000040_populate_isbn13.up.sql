-- Strip non-digits from isbn10 and store as isbn13 (simple cleaning)
UPDATE public.book_editions
SET isbn13 = regexp_replace(isbn10, '[^0-9]', '', 'g')
WHERE isbn10 IS NOT NULL
  AND isbn10 != ''
  AND LENGTH(regexp_replace(isbn10, '[^0-9]', '', 'g')) IN (10, 13);
