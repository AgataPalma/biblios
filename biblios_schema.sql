--
-- PostgreSQL database dump
--

-- Dumped from database version 16.9 (Debian 16.9-1.pgdg120+1)
-- Dumped by pg_dump version 16.9 (Debian 16.9-1.pgdg120+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


--
-- Name: books_search_vector(text, text, text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.books_search_vector(p_title text, p_description text, p_authors text) RETURNS tsvector
    LANGUAGE plpgsql IMMUTABLE
    AS $$
BEGIN
    RETURN (
        setweight(to_tsvector('english', coalesce(p_title, '')), 'A') ||
        setweight(to_tsvector('english', coalesce(p_authors, '')), 'B') ||
        setweight(to_tsvector('english', coalesce(p_description, '')), 'C')
    );
END;
$$;


--
-- Name: books_search_vector_trigger(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.books_search_vector_trigger() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_authors TEXT;
BEGIN
    SELECT string_agg(c.name, ' ')
    INTO v_authors
    FROM book_contributors bc
             JOIN contributors c ON c.id = bc.contributor_id
    WHERE bc.book_id = NEW.id;

    NEW.search_vector := public.books_search_vector(NEW.title, NEW.description, v_authors);
    RETURN NEW;
END;
$$;


--
-- Name: update_updated_at(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_updated_at() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: awards; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.awards (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    description text
);


--
-- Name: book_awards; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.book_awards (
    book_id uuid NOT NULL,
    award_id uuid NOT NULL,
    year integer NOT NULL,
    category character varying(255),
    result character varying(20),
    CONSTRAINT book_awards_result_check CHECK (((result)::text = ANY ((ARRAY['winner'::character varying, 'nominee'::character varying])::text[])))
);


--
-- Name: book_contributors; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.book_contributors (
    book_id uuid NOT NULL,
    contributor_id uuid NOT NULL,
    role character varying(20) NOT NULL,
    CONSTRAINT book_contributors_role_check CHECK (((role)::text = ANY ((ARRAY['author'::character varying, 'co_author'::character varying])::text[])))
);


--
-- Name: book_copies; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.book_copies (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    edition_id uuid NOT NULL,
    owner_id uuid NOT NULL,
    condition character varying(10),
    deleted_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    reading_status character varying(20) DEFAULT 'want_to_read'::character varying NOT NULL,
    current_page integer,
    started_reading_at timestamp with time zone,
    finished_reading_at timestamp with time zone,
    owned_by_user boolean DEFAULT true NOT NULL,
    borrowed_from uuid,
    location character varying(200),
    reread_count integer DEFAULT 0 NOT NULL,
    personal_notes text,
    CONSTRAINT book_copies_condition_check CHECK (((condition)::text = ANY ((ARRAY['new'::character varying, 'good'::character varying, 'fair'::character varying, 'poor'::character varying])::text[]))),
    CONSTRAINT book_copies_reading_status_check CHECK (((reading_status)::text = ANY ((ARRAY['want_to_read'::character varying, 'reading'::character varying, 'read'::character varying, 'did_not_finish'::character varying])::text[])))
);


--
-- Name: book_editions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.book_editions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    book_id uuid NOT NULL,
    format character varying(20) NOT NULL,
    isbn10 character varying(20),
    asin character varying(20),
    language character varying(10) NOT NULL,
    publisher character varying(255),
    edition character varying(50),
    published_at date,
    page_count integer,
    file_format character varying(10),
    duration_minutes integer,
    audio_format character varying(10),
    deleted_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    status character varying(10) DEFAULT 'pending'::character varying NOT NULL,
    rejection_reason text,
    description text,
    cover_url text,
    isbn13 character varying(17),
    title character varying(500) NOT NULL,
    original_title character varying(500),
    CONSTRAINT book_editions_audio_format_check CHECK (((audio_format)::text = ANY ((ARRAY['MP3'::character varying, 'AAC'::character varying, 'WMA'::character varying, 'FLAC'::character varying])::text[]))),
    CONSTRAINT book_editions_file_format_check CHECK (((file_format)::text = ANY ((ARRAY['EPUB'::character varying, 'PDF'::character varying, 'MOBI'::character varying, 'AZW3'::character varying])::text[]))),
    CONSTRAINT book_editions_format_check CHECK (((format)::text = ANY ((ARRAY['hardcover'::character varying, 'paperback'::character varying, 'ebook'::character varying, 'audiobook'::character varying, 'graphic_novel'::character varying])::text[]))),
    CONSTRAINT book_editions_isbn13_format_check CHECK (((isbn13 IS NULL) OR ((isbn13)::text ~ '^[0-9]{13}$|^[0-9]{1,5}-?[0-9]{1,7}-?[0-9]{1,6}-?[0-9]$'::text))),
    CONSTRAINT book_editions_status_check CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'approved'::character varying, 'rejected'::character varying])::text[])))
);


--
-- Name: book_genres; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.book_genres (
    book_id uuid NOT NULL,
    genre_id uuid NOT NULL
);


--
-- Name: book_moods; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.book_moods (
    book_id uuid NOT NULL,
    mood_id uuid NOT NULL
);


--
-- Name: books; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.books (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    title character varying(500) NOT NULL,
    deleted_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    status character varying(10) DEFAULT 'pending'::character varying NOT NULL,
    rejection_reason text,
    search_vector tsvector,
    series_id uuid,
    series_position numeric(5,1),
    description text,
    CONSTRAINT books_status_check CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'approved'::character varying, 'rejected'::character varying])::text[])))
);


--
-- Name: collection_books; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.collection_books (
    collection_id uuid NOT NULL,
    book_copy_id uuid NOT NULL,
    added_by uuid NOT NULL,
    added_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: collections; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.collections (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    library_id uuid NOT NULL,
    created_by uuid NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    cover_colour character varying(7),
    is_public boolean DEFAULT false NOT NULL,
    deleted_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    is_collaborative boolean DEFAULT false NOT NULL
);


--
-- Name: contributor_awards; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.contributor_awards (
    contributor_id uuid NOT NULL,
    award_id uuid NOT NULL,
    year integer NOT NULL,
    category character varying(255),
    result character varying(20),
    CONSTRAINT contributor_awards_result_check CHECK (((result)::text = ANY ((ARRAY['winner'::character varying, 'nominee'::character varying])::text[])))
);


--
-- Name: contributors; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.contributors (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    bio text,
    born_date date,
    died_date date,
    photo_url text,
    website text,
    nationality character varying(100),
    status character varying(10) DEFAULT 'pending'::character varying NOT NULL,
    rejection_reason text,
    deleted_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT contributors_status_check CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'approved'::character varying, 'rejected'::character varying])::text[])))
);


--
-- Name: edition_contributors; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.edition_contributors (
    edition_id uuid NOT NULL,
    contributor_id uuid NOT NULL,
    role character varying(20) NOT NULL,
    CONSTRAINT edition_contributors_role_check CHECK (((role)::text = ANY ((ARRAY['narrator'::character varying, 'translator'::character varying, 'illustrator'::character varying, 'editor'::character varying])::text[])))
);


--
-- Name: email_queue; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.email_queue (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    to_email character varying(255) NOT NULL,
    subject character varying(255) NOT NULL,
    body text NOT NULL,
    status character varying(10) DEFAULT 'pending'::character varying NOT NULL,
    attempts integer DEFAULT 0 NOT NULL,
    last_error text,
    sent_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT email_queue_status_check CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'sent'::character varying, 'failed'::character varying])::text[])))
);


--
-- Name: genres; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.genres (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(100) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    status character varying(10) DEFAULT 'pending'::character varying NOT NULL,
    rejection_reason text,
    CONSTRAINT genres_status_check CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'approved'::character varying, 'rejected'::character varying])::text[])))
);


--
-- Name: import_jobs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.import_jobs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    status character varying(20) DEFAULT 'pending'::character varying NOT NULL,
    source character varying(50) DEFAULT 'goodreads'::character varying NOT NULL,
    total_rows integer,
    processed integer DEFAULT 0 NOT NULL,
    imported integer DEFAULT 0 NOT NULL,
    skipped integer DEFAULT 0 NOT NULL,
    failed_rows integer DEFAULT 0 NOT NULL,
    error_message text,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT import_jobs_status_check CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'processing'::character varying, 'completed'::character varying, 'failed'::character varying])::text[])))
);


--
-- Name: libraries; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.libraries (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    owner_id uuid NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    is_cooperative boolean DEFAULT false NOT NULL,
    visibility character varying(20) DEFAULT 'private'::character varying NOT NULL,
    deleted_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT libraries_visibility_check CHECK (((visibility)::text = ANY ((ARRAY['private'::character varying, 'semi_public'::character varying, 'public'::character varying])::text[])))
);


--
-- Name: library_book_copies; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.library_book_copies (
    library_id uuid NOT NULL,
    book_copy_id uuid NOT NULL,
    added_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: library_invitations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.library_invitations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    library_id uuid NOT NULL,
    invited_by uuid NOT NULL,
    invited_user_id uuid,
    invited_email character varying(255) NOT NULL,
    token character varying(128) NOT NULL,
    status character varying(20) DEFAULT 'pending'::character varying NOT NULL,
    accepted_at timestamp with time zone,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT library_invitations_status_check CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'accepted'::character varying, 'declined'::character varying, 'expired'::character varying, 'revoked'::character varying])::text[])))
);


--
-- Name: library_members; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.library_members (
    library_id uuid NOT NULL,
    user_id uuid NOT NULL,
    joined_at timestamp with time zone DEFAULT now() NOT NULL,
    is_owner boolean DEFAULT false NOT NULL,
    can_view boolean DEFAULT true NOT NULL,
    can_add boolean DEFAULT false NOT NULL,
    can_remove boolean DEFAULT false NOT NULL,
    can_edit boolean DEFAULT false NOT NULL,
    can_invite boolean DEFAULT false NOT NULL,
    can_manage_members boolean DEFAULT false NOT NULL,
    CONSTRAINT owner_has_all_permissions CHECK (((is_owner = false) OR ((can_view = true) AND (can_add = true) AND (can_remove = true) AND (can_edit = true) AND (can_invite = true) AND (can_manage_members = true))))
);


--
-- Name: moderation_log; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.moderation_log (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    moderator_id uuid NOT NULL,
    entity_type character varying(50) NOT NULL,
    entity_id uuid NOT NULL,
    action character varying(20) NOT NULL,
    before jsonb,
    after jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT moderation_log_action_check CHECK (((action)::text = ANY ((ARRAY['approved'::character varying, 'rejected'::character varying, 'edited'::character varying])::text[])))
);


--
-- Name: moods; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.moods (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(50) NOT NULL,
    status character varying(10) DEFAULT 'approved'::character varying NOT NULL,
    rejection_reason text,
    CONSTRAINT moods_status_check CHECK (((status)::text = ANY (ARRAY[('pending'::character varying)::text, ('approved'::character varying)::text, ('rejected'::character varying)::text])))
);


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    type character varying(50) NOT NULL,
    title character varying(255) NOT NULL,
    body text NOT NULL,
    read_at timestamp with time zone,
    data jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: quotes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.quotes (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    book_id uuid NOT NULL,
    user_id uuid,
    body text NOT NULL,
    page_number integer,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: reading_challenges; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reading_challenges (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    year integer NOT NULL,
    goal integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: reading_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reading_sessions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    copy_id uuid NOT NULL,
    logged_date date NOT NULL,
    pages_read integer,
    progress_pct numeric(5,2),
    note text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT reading_sessions_progress_pct_check CHECK (((progress_pct >= (0)::numeric) AND (progress_pct <= (100)::numeric)))
);


--
-- Name: review_likes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.review_likes (
    review_id uuid NOT NULL,
    user_id uuid NOT NULL,
    liked_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: reviews; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reviews (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    book_id uuid NOT NULL,
    user_id uuid,
    rating numeric(2,1) NOT NULL,
    body text,
    is_public boolean DEFAULT true NOT NULL,
    deleted_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    like_count integer DEFAULT 0 NOT NULL,
    CONSTRAINT reviews_body_check CHECK ((char_length(body) <= 5000)),
    CONSTRAINT reviews_rating_check CHECK (((rating >= 0.5) AND (rating <= 5.0) AND ((rating * (2)::numeric) = floor((rating * (2)::numeric)))))
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);


--
-- Name: series; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.series (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    status character varying(10) DEFAULT 'pending'::character varying NOT NULL,
    rejection_reason text,
    deleted_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT series_status_check CHECK (((status)::text = ANY (ARRAY[('pending'::character varying)::text, ('approved'::character varying)::text, ('rejected'::character varying)::text])))
);


--
-- Name: shelf_books; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.shelf_books (
    shelf_id uuid NOT NULL,
    copy_id uuid NOT NULL,
    added_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: shelves; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.shelves (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    name character varying(100) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: submissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.submissions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    submitted_by uuid NOT NULL,
    status character varying(10) DEFAULT 'pending'::character varying NOT NULL,
    rejection_reason text,
    reviewed_by uuid,
    reviewed_at timestamp with time zone,
    book_id uuid,
    edition_id uuid,
    copy_id uuid,
    deleted_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    catalogue_only boolean DEFAULT false NOT NULL,
    contributor_id uuid,
    CONSTRAINT submissions_status_check CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'approved'::character varying, 'rejected'::character varying])::text[])))
);


--
-- Name: user_follows; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_follows (
    follower_id uuid NOT NULL,
    following_id uuid NOT NULL,
    followed_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT user_follows_check CHECK ((follower_id <> following_id))
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    email character varying(255) NOT NULL,
    username character varying(50) NOT NULL,
    password_hash text NOT NULL,
    is_admin boolean DEFAULT false NOT NULL,
    deleted_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    role character varying(10) DEFAULT 'user'::character varying NOT NULL,
    theme character varying(50) DEFAULT 'default-light'::character varying NOT NULL,
    bio text,
    avatar_url text,
    CONSTRAINT users_admin_role_sync CHECK (((is_admin = false) OR ((role)::text = 'admin'::text))),
    CONSTRAINT users_role_check CHECK (((role)::text = ANY ((ARRAY['user'::character varying, 'moderator'::character varying, 'admin'::character varying])::text[])))
);


--
-- Name: awards awards_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.awards
    ADD CONSTRAINT awards_name_key UNIQUE (name);


--
-- Name: awards awards_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.awards
    ADD CONSTRAINT awards_pkey PRIMARY KEY (id);


--
-- Name: book_awards book_awards_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.book_awards
    ADD CONSTRAINT book_awards_pkey PRIMARY KEY (book_id, award_id, year);


--
-- Name: book_contributors book_contributors_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.book_contributors
    ADD CONSTRAINT book_contributors_pkey PRIMARY KEY (book_id, contributor_id, role);


--
-- Name: book_copies book_copies_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.book_copies
    ADD CONSTRAINT book_copies_pkey PRIMARY KEY (id);


--
-- Name: book_editions book_editions_isbn13_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.book_editions
    ADD CONSTRAINT book_editions_isbn13_unique UNIQUE (isbn13);


--
-- Name: book_editions book_editions_isbn_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.book_editions
    ADD CONSTRAINT book_editions_isbn_unique UNIQUE (isbn10);


--
-- Name: book_editions book_editions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.book_editions
    ADD CONSTRAINT book_editions_pkey PRIMARY KEY (id);


--
-- Name: book_genres book_genres_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.book_genres
    ADD CONSTRAINT book_genres_pkey PRIMARY KEY (book_id, genre_id);


--
-- Name: book_moods book_moods_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.book_moods
    ADD CONSTRAINT book_moods_pkey PRIMARY KEY (book_id, mood_id);


--
-- Name: books books_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.books
    ADD CONSTRAINT books_pkey PRIMARY KEY (id);


--
-- Name: collection_books collection_books_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collection_books
    ADD CONSTRAINT collection_books_pkey PRIMARY KEY (collection_id, book_copy_id);


--
-- Name: collections collections_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collections
    ADD CONSTRAINT collections_pkey PRIMARY KEY (id);


--
-- Name: contributor_awards contributor_awards_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contributor_awards
    ADD CONSTRAINT contributor_awards_pkey PRIMARY KEY (contributor_id, award_id, year);


--
-- Name: contributors contributors_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contributors
    ADD CONSTRAINT contributors_pkey PRIMARY KEY (id);


--
-- Name: edition_contributors edition_contributors_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.edition_contributors
    ADD CONSTRAINT edition_contributors_pkey PRIMARY KEY (edition_id, contributor_id, role);


--
-- Name: email_queue email_queue_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_queue
    ADD CONSTRAINT email_queue_pkey PRIMARY KEY (id);


--
-- Name: genres genres_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.genres
    ADD CONSTRAINT genres_name_key UNIQUE (name);


--
-- Name: genres genres_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.genres
    ADD CONSTRAINT genres_pkey PRIMARY KEY (id);


--
-- Name: import_jobs import_jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.import_jobs
    ADD CONSTRAINT import_jobs_pkey PRIMARY KEY (id);


--
-- Name: libraries libraries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.libraries
    ADD CONSTRAINT libraries_pkey PRIMARY KEY (id);


--
-- Name: library_book_copies library_book_copies_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.library_book_copies
    ADD CONSTRAINT library_book_copies_pkey PRIMARY KEY (library_id, book_copy_id);


--
-- Name: library_invitations library_invitations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.library_invitations
    ADD CONSTRAINT library_invitations_pkey PRIMARY KEY (id);


--
-- Name: library_invitations library_invitations_token_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.library_invitations
    ADD CONSTRAINT library_invitations_token_key UNIQUE (token);


--
-- Name: library_members library_members_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.library_members
    ADD CONSTRAINT library_members_pkey PRIMARY KEY (library_id, user_id);


--
-- Name: moderation_log moderation_log_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.moderation_log
    ADD CONSTRAINT moderation_log_pkey PRIMARY KEY (id);


--
-- Name: moods moods_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.moods
    ADD CONSTRAINT moods_name_key UNIQUE (name);


--
-- Name: moods moods_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.moods
    ADD CONSTRAINT moods_pkey PRIMARY KEY (id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: quotes quotes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.quotes
    ADD CONSTRAINT quotes_pkey PRIMARY KEY (id);


--
-- Name: reading_challenges reading_challenges_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reading_challenges
    ADD CONSTRAINT reading_challenges_pkey PRIMARY KEY (id);


--
-- Name: reading_challenges reading_challenges_user_id_year_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reading_challenges
    ADD CONSTRAINT reading_challenges_user_id_year_key UNIQUE (user_id, year);


--
-- Name: reading_sessions reading_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reading_sessions
    ADD CONSTRAINT reading_sessions_pkey PRIMARY KEY (id);


--
-- Name: review_likes review_likes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.review_likes
    ADD CONSTRAINT review_likes_pkey PRIMARY KEY (review_id, user_id);


--
-- Name: reviews reviews_book_id_user_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_book_id_user_id_key UNIQUE (book_id, user_id);


--
-- Name: reviews reviews_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: series series_name_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.series
    ADD CONSTRAINT series_name_unique UNIQUE (name);


--
-- Name: series series_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.series
    ADD CONSTRAINT series_pkey PRIMARY KEY (id);


--
-- Name: shelf_books shelf_books_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.shelf_books
    ADD CONSTRAINT shelf_books_pkey PRIMARY KEY (shelf_id, copy_id);


--
-- Name: shelves shelves_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.shelves
    ADD CONSTRAINT shelves_pkey PRIMARY KEY (id);


--
-- Name: shelves shelves_user_id_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.shelves
    ADD CONSTRAINT shelves_user_id_name_key UNIQUE (user_id, name);


--
-- Name: submissions submissions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.submissions
    ADD CONSTRAINT submissions_pkey PRIMARY KEY (id);


--
-- Name: user_follows user_follows_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_follows
    ADD CONSTRAINT user_follows_pkey PRIMARY KEY (follower_id, following_id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_username_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- Name: idx_book_awards_award; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_book_awards_award ON public.book_awards USING btree (award_id);


--
-- Name: idx_book_awards_book; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_book_awards_book ON public.book_awards USING btree (book_id);


--
-- Name: idx_book_contributors_book; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_book_contributors_book ON public.book_contributors USING btree (book_id);


--
-- Name: idx_book_contributors_contributor; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_book_contributors_contributor ON public.book_contributors USING btree (contributor_id);


--
-- Name: idx_book_copies_edition; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_book_copies_edition ON public.book_copies USING btree (edition_id);


--
-- Name: idx_book_copies_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_book_copies_owner ON public.book_copies USING btree (owner_id);


--
-- Name: idx_book_copies_owner_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_book_copies_owner_status ON public.book_copies USING btree (owner_id, reading_status) WHERE (deleted_at IS NULL);


--
-- Name: idx_book_copies_reading_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_book_copies_reading_status ON public.book_copies USING btree (reading_status) WHERE (deleted_at IS NULL);


--
-- Name: idx_book_editions_book; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_book_editions_book ON public.book_editions USING btree (book_id);


--
-- Name: idx_book_editions_dedupe; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_book_editions_dedupe ON public.book_editions USING btree (lower((original_title)::text), lower((language)::text));


--
-- Name: idx_book_editions_format; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_book_editions_format ON public.book_editions USING btree (format);


--
-- Name: idx_book_editions_isbn; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_book_editions_isbn ON public.book_editions USING btree (isbn10);


--
-- Name: idx_book_editions_isbn10; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_book_editions_isbn10 ON public.book_editions USING btree (isbn10);


--
-- Name: idx_book_editions_isbn13; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_book_editions_isbn13 ON public.book_editions USING btree (isbn13);


--
-- Name: idx_book_editions_original_title; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_book_editions_original_title ON public.book_editions USING btree (lower((original_title)::text));


--
-- Name: idx_book_editions_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_book_editions_status ON public.book_editions USING btree (status) WHERE (deleted_at IS NULL);


--
-- Name: idx_book_editions_title; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_book_editions_title ON public.book_editions USING btree (lower((title)::text));


--
-- Name: idx_book_editions_titles_dual_fts; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_book_editions_titles_dual_fts ON public.book_editions USING gin (((setweight(to_tsvector('portuguese'::regconfig, (((COALESCE(title, ''::character varying))::text || ' '::text) || (COALESCE(original_title, ''::character varying))::text)), 'A'::"char") || setweight(to_tsvector('english'::regconfig, (((COALESCE(title, ''::character varying))::text || ' '::text) || (COALESCE(original_title, ''::character varying))::text)), 'B'::"char"))));


--
-- Name: idx_book_genres_book; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_book_genres_book ON public.book_genres USING btree (book_id);


--
-- Name: idx_book_genres_genre; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_book_genres_genre ON public.book_genres USING btree (genre_id);


--
-- Name: idx_book_moods_book; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_book_moods_book ON public.book_moods USING btree (book_id);


--
-- Name: idx_books_search_vector; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_books_search_vector ON public.books USING gin (search_vector);


--
-- Name: idx_books_series; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_books_series ON public.books USING btree (series_id) WHERE (series_id IS NOT NULL);


--
-- Name: idx_books_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_books_status ON public.books USING btree (status) WHERE (deleted_at IS NULL);


--
-- Name: idx_books_title; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_books_title ON public.books USING btree (title);


--
-- Name: idx_collection_books_collection; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_collection_books_collection ON public.collection_books USING btree (collection_id);


--
-- Name: idx_collection_books_copy; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_collection_books_copy ON public.collection_books USING btree (book_copy_id);


--
-- Name: idx_collections_created_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_collections_created_by ON public.collections USING btree (created_by);


--
-- Name: idx_collections_library; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_collections_library ON public.collections USING btree (library_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_contributor_awards_award; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_contributor_awards_award ON public.contributor_awards USING btree (award_id);


--
-- Name: idx_contributor_awards_contributor; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_contributor_awards_contributor ON public.contributor_awards USING btree (contributor_id);


--
-- Name: idx_contributors_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_contributors_name ON public.contributors USING btree (name);


--
-- Name: idx_contributors_name_fts; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_contributors_name_fts ON public.contributors USING gin (to_tsvector('english'::regconfig, (name)::text));


--
-- Name: idx_edition_contributors_contributor; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_edition_contributors_contributor ON public.edition_contributors USING btree (contributor_id);


--
-- Name: idx_edition_contributors_edition; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_edition_contributors_edition ON public.edition_contributors USING btree (edition_id);


--
-- Name: idx_email_queue_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_email_queue_status ON public.email_queue USING btree (status);


--
-- Name: idx_email_queue_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_email_queue_user ON public.email_queue USING btree (user_id);


--
-- Name: idx_import_jobs_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_import_jobs_status ON public.import_jobs USING btree (status) WHERE ((status)::text = ANY ((ARRAY['pending'::character varying, 'processing'::character varying])::text[]));


--
-- Name: idx_import_jobs_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_import_jobs_user ON public.import_jobs USING btree (user_id);


--
-- Name: idx_libraries_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_libraries_owner ON public.libraries USING btree (owner_id);


--
-- Name: idx_libraries_visibility; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_libraries_visibility ON public.libraries USING btree (visibility) WHERE (deleted_at IS NULL);


--
-- Name: idx_library_book_copies_copy; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_library_book_copies_copy ON public.library_book_copies USING btree (book_copy_id);


--
-- Name: idx_library_book_copies_library; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_library_book_copies_library ON public.library_book_copies USING btree (library_id);


--
-- Name: idx_library_invitations_email; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_library_invitations_email ON public.library_invitations USING btree (invited_email);


--
-- Name: idx_library_invitations_library; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_library_invitations_library ON public.library_invitations USING btree (library_id);


--
-- Name: idx_library_invitations_token; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_library_invitations_token ON public.library_invitations USING btree (token);


--
-- Name: idx_library_members_library; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_library_members_library ON public.library_members USING btree (library_id);


--
-- Name: idx_library_members_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_library_members_user ON public.library_members USING btree (user_id);


--
-- Name: idx_moderation_log_action; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_moderation_log_action ON public.moderation_log USING btree (action);


--
-- Name: idx_moderation_log_entity; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_moderation_log_entity ON public.moderation_log USING btree (entity_type, entity_id);


--
-- Name: idx_moderation_log_moderator; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_moderation_log_moderator ON public.moderation_log USING btree (moderator_id);


--
-- Name: idx_moods_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_moods_status ON public.moods USING btree (status);


--
-- Name: idx_notifications_read_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_read_at ON public.notifications USING btree (read_at);


--
-- Name: idx_notifications_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_type ON public.notifications USING btree (type);


--
-- Name: idx_notifications_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_user ON public.notifications USING btree (user_id);


--
-- Name: idx_notifications_user_read; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_notifications_user_read ON public.notifications USING btree (user_id, read_at);


--
-- Name: idx_quotes_book; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_quotes_book ON public.quotes USING btree (book_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_quotes_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_quotes_user ON public.quotes USING btree (user_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_reading_challenges_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_reading_challenges_user ON public.reading_challenges USING btree (user_id);


--
-- Name: idx_reading_sessions_copy; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_reading_sessions_copy ON public.reading_sessions USING btree (copy_id);


--
-- Name: idx_reading_sessions_user_date; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_reading_sessions_user_date ON public.reading_sessions USING btree (user_id, logged_date DESC);


--
-- Name: idx_review_likes_review; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_review_likes_review ON public.review_likes USING btree (review_id);


--
-- Name: idx_review_likes_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_review_likes_user ON public.review_likes USING btree (user_id);


--
-- Name: idx_reviews_book; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_reviews_book ON public.reviews USING btree (book_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_reviews_public; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_reviews_public ON public.reviews USING btree (book_id, rating) WHERE ((is_public = true) AND (deleted_at IS NULL));


--
-- Name: idx_reviews_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_reviews_user ON public.reviews USING btree (user_id) WHERE (deleted_at IS NULL);


--
-- Name: idx_series_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_series_name ON public.series USING btree (name);


--
-- Name: idx_series_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_series_status ON public.series USING btree (status) WHERE (deleted_at IS NULL);


--
-- Name: idx_shelf_books_copy; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_shelf_books_copy ON public.shelf_books USING btree (copy_id);


--
-- Name: idx_shelves_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_shelves_user ON public.shelves USING btree (user_id);


--
-- Name: idx_submissions_contributor; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_submissions_contributor ON public.submissions USING btree (contributor_id);


--
-- Name: idx_submissions_reviewed_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_submissions_reviewed_by ON public.submissions USING btree (reviewed_by);


--
-- Name: idx_submissions_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_submissions_status ON public.submissions USING btree (status);


--
-- Name: idx_submissions_submitted_by; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_submissions_submitted_by ON public.submissions USING btree (submitted_by);


--
-- Name: idx_user_follows_follower; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_follows_follower ON public.user_follows USING btree (follower_id);


--
-- Name: idx_user_follows_following; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_follows_following ON public.user_follows USING btree (following_id);


--
-- Name: idx_users_email; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_email ON public.users USING btree (email);


--
-- Name: idx_users_username; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_username ON public.users USING btree (username);


--
-- Name: book_copies book_copies_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER book_copies_updated_at BEFORE UPDATE ON public.book_copies FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();


--
-- Name: book_editions book_editions_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER book_editions_updated_at BEFORE UPDATE ON public.book_editions FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();


--
-- Name: books books_search_vector_update; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER books_search_vector_update BEFORE INSERT OR UPDATE ON public.books FOR EACH ROW EXECUTE FUNCTION public.books_search_vector_trigger();


--
-- Name: books books_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER books_updated_at BEFORE UPDATE ON public.books FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();


--
-- Name: collections collections_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER collections_updated_at BEFORE UPDATE ON public.collections FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();


--
-- Name: contributors contributors_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER contributors_updated_at BEFORE UPDATE ON public.contributors FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();


--
-- Name: email_queue email_queue_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER email_queue_updated_at BEFORE UPDATE ON public.email_queue FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();


--
-- Name: import_jobs import_jobs_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER import_jobs_updated_at BEFORE UPDATE ON public.import_jobs FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();


--
-- Name: libraries libraries_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER libraries_updated_at BEFORE UPDATE ON public.libraries FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();


--
-- Name: reviews reviews_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER reviews_updated_at BEFORE UPDATE ON public.reviews FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();


--
-- Name: series series_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER series_updated_at BEFORE UPDATE ON public.series FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();


--
-- Name: submissions submissions_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER submissions_updated_at BEFORE UPDATE ON public.submissions FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();


--
-- Name: users users_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER users_updated_at BEFORE UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION public.update_updated_at();


--
-- Name: book_awards book_awards_award_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.book_awards
    ADD CONSTRAINT book_awards_award_id_fkey FOREIGN KEY (award_id) REFERENCES public.awards(id) ON DELETE CASCADE;


--
-- Name: book_awards book_awards_book_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.book_awards
    ADD CONSTRAINT book_awards_book_id_fkey FOREIGN KEY (book_id) REFERENCES public.books(id) ON DELETE CASCADE;


--
-- Name: book_contributors book_contributors_book_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.book_contributors
    ADD CONSTRAINT book_contributors_book_id_fkey FOREIGN KEY (book_id) REFERENCES public.books(id) ON DELETE CASCADE;


--
-- Name: book_contributors book_contributors_contributor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.book_contributors
    ADD CONSTRAINT book_contributors_contributor_id_fkey FOREIGN KEY (contributor_id) REFERENCES public.contributors(id) ON DELETE CASCADE;


--
-- Name: book_copies book_copies_borrowed_from_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.book_copies
    ADD CONSTRAINT book_copies_borrowed_from_fkey FOREIGN KEY (borrowed_from) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: book_copies book_copies_edition_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.book_copies
    ADD CONSTRAINT book_copies_edition_id_fkey FOREIGN KEY (edition_id) REFERENCES public.book_editions(id) ON DELETE CASCADE;


--
-- Name: book_copies book_copies_owner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.book_copies
    ADD CONSTRAINT book_copies_owner_id_fkey FOREIGN KEY (owner_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: book_editions book_editions_book_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.book_editions
    ADD CONSTRAINT book_editions_book_id_fkey FOREIGN KEY (book_id) REFERENCES public.books(id) ON DELETE CASCADE;


--
-- Name: book_genres book_genres_book_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.book_genres
    ADD CONSTRAINT book_genres_book_id_fkey FOREIGN KEY (book_id) REFERENCES public.books(id) ON DELETE CASCADE;


--
-- Name: book_genres book_genres_genre_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.book_genres
    ADD CONSTRAINT book_genres_genre_id_fkey FOREIGN KEY (genre_id) REFERENCES public.genres(id) ON DELETE CASCADE;


--
-- Name: book_moods book_moods_book_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.book_moods
    ADD CONSTRAINT book_moods_book_id_fkey FOREIGN KEY (book_id) REFERENCES public.books(id) ON DELETE CASCADE;


--
-- Name: book_moods book_moods_mood_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.book_moods
    ADD CONSTRAINT book_moods_mood_id_fkey FOREIGN KEY (mood_id) REFERENCES public.moods(id) ON DELETE CASCADE;


--
-- Name: books books_series_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.books
    ADD CONSTRAINT books_series_id_fkey FOREIGN KEY (series_id) REFERENCES public.series(id) ON DELETE SET NULL;


--
-- Name: collection_books collection_books_added_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collection_books
    ADD CONSTRAINT collection_books_added_by_fkey FOREIGN KEY (added_by) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: collection_books collection_books_book_copy_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collection_books
    ADD CONSTRAINT collection_books_book_copy_id_fkey FOREIGN KEY (book_copy_id) REFERENCES public.book_copies(id) ON DELETE CASCADE;


--
-- Name: collection_books collection_books_collection_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collection_books
    ADD CONSTRAINT collection_books_collection_id_fkey FOREIGN KEY (collection_id) REFERENCES public.collections(id) ON DELETE CASCADE;


--
-- Name: collections collections_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collections
    ADD CONSTRAINT collections_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: collections collections_library_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collections
    ADD CONSTRAINT collections_library_id_fkey FOREIGN KEY (library_id) REFERENCES public.libraries(id) ON DELETE CASCADE;


--
-- Name: contributor_awards contributor_awards_award_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contributor_awards
    ADD CONSTRAINT contributor_awards_award_id_fkey FOREIGN KEY (award_id) REFERENCES public.awards(id) ON DELETE CASCADE;


--
-- Name: contributor_awards contributor_awards_contributor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contributor_awards
    ADD CONSTRAINT contributor_awards_contributor_id_fkey FOREIGN KEY (contributor_id) REFERENCES public.contributors(id) ON DELETE CASCADE;


--
-- Name: edition_contributors edition_contributors_contributor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.edition_contributors
    ADD CONSTRAINT edition_contributors_contributor_id_fkey FOREIGN KEY (contributor_id) REFERENCES public.contributors(id) ON DELETE CASCADE;


--
-- Name: edition_contributors edition_contributors_edition_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.edition_contributors
    ADD CONSTRAINT edition_contributors_edition_id_fkey FOREIGN KEY (edition_id) REFERENCES public.book_editions(id) ON DELETE CASCADE;


--
-- Name: email_queue email_queue_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_queue
    ADD CONSTRAINT email_queue_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: import_jobs import_jobs_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.import_jobs
    ADD CONSTRAINT import_jobs_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: libraries libraries_owner_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.libraries
    ADD CONSTRAINT libraries_owner_id_fkey FOREIGN KEY (owner_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: library_book_copies library_book_copies_book_copy_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.library_book_copies
    ADD CONSTRAINT library_book_copies_book_copy_id_fkey FOREIGN KEY (book_copy_id) REFERENCES public.book_copies(id) ON DELETE CASCADE;


--
-- Name: library_book_copies library_book_copies_library_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.library_book_copies
    ADD CONSTRAINT library_book_copies_library_id_fkey FOREIGN KEY (library_id) REFERENCES public.libraries(id) ON DELETE CASCADE;


--
-- Name: library_invitations library_invitations_invited_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.library_invitations
    ADD CONSTRAINT library_invitations_invited_by_fkey FOREIGN KEY (invited_by) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: library_invitations library_invitations_invited_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.library_invitations
    ADD CONSTRAINT library_invitations_invited_user_id_fkey FOREIGN KEY (invited_user_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: library_invitations library_invitations_library_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.library_invitations
    ADD CONSTRAINT library_invitations_library_id_fkey FOREIGN KEY (library_id) REFERENCES public.libraries(id) ON DELETE CASCADE;


--
-- Name: library_members library_members_library_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.library_members
    ADD CONSTRAINT library_members_library_id_fkey FOREIGN KEY (library_id) REFERENCES public.libraries(id) ON DELETE CASCADE;


--
-- Name: library_members library_members_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.library_members
    ADD CONSTRAINT library_members_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: moderation_log moderation_log_moderator_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.moderation_log
    ADD CONSTRAINT moderation_log_moderator_id_fkey FOREIGN KEY (moderator_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: notifications notifications_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: quotes quotes_book_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.quotes
    ADD CONSTRAINT quotes_book_id_fkey FOREIGN KEY (book_id) REFERENCES public.books(id) ON DELETE CASCADE;


--
-- Name: quotes quotes_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.quotes
    ADD CONSTRAINT quotes_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: reading_challenges reading_challenges_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reading_challenges
    ADD CONSTRAINT reading_challenges_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: reading_sessions reading_sessions_copy_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reading_sessions
    ADD CONSTRAINT reading_sessions_copy_id_fkey FOREIGN KEY (copy_id) REFERENCES public.book_copies(id) ON DELETE CASCADE;


--
-- Name: reading_sessions reading_sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reading_sessions
    ADD CONSTRAINT reading_sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: review_likes review_likes_review_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.review_likes
    ADD CONSTRAINT review_likes_review_id_fkey FOREIGN KEY (review_id) REFERENCES public.reviews(id) ON DELETE CASCADE;


--
-- Name: review_likes review_likes_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.review_likes
    ADD CONSTRAINT review_likes_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: reviews reviews_book_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_book_id_fkey FOREIGN KEY (book_id) REFERENCES public.books(id) ON DELETE CASCADE;


--
-- Name: reviews reviews_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: shelf_books shelf_books_copy_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.shelf_books
    ADD CONSTRAINT shelf_books_copy_id_fkey FOREIGN KEY (copy_id) REFERENCES public.book_copies(id) ON DELETE CASCADE;


--
-- Name: shelf_books shelf_books_shelf_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.shelf_books
    ADD CONSTRAINT shelf_books_shelf_id_fkey FOREIGN KEY (shelf_id) REFERENCES public.shelves(id) ON DELETE CASCADE;


--
-- Name: shelves shelves_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.shelves
    ADD CONSTRAINT shelves_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: submissions submissions_book_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.submissions
    ADD CONSTRAINT submissions_book_id_fkey FOREIGN KEY (book_id) REFERENCES public.books(id) ON DELETE SET NULL;


--
-- Name: submissions submissions_contributor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.submissions
    ADD CONSTRAINT submissions_contributor_id_fkey FOREIGN KEY (contributor_id) REFERENCES public.contributors(id) ON DELETE SET NULL;


--
-- Name: submissions submissions_copy_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.submissions
    ADD CONSTRAINT submissions_copy_id_fkey FOREIGN KEY (copy_id) REFERENCES public.book_copies(id) ON DELETE SET NULL;


--
-- Name: submissions submissions_edition_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.submissions
    ADD CONSTRAINT submissions_edition_id_fkey FOREIGN KEY (edition_id) REFERENCES public.book_editions(id) ON DELETE SET NULL;


--
-- Name: submissions submissions_reviewed_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.submissions
    ADD CONSTRAINT submissions_reviewed_by_fkey FOREIGN KEY (reviewed_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: submissions submissions_submitted_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.submissions
    ADD CONSTRAINT submissions_submitted_by_fkey FOREIGN KEY (submitted_by) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_follows user_follows_follower_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_follows
    ADD CONSTRAINT user_follows_follower_id_fkey FOREIGN KEY (follower_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_follows user_follows_following_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_follows
    ADD CONSTRAINT user_follows_following_id_fkey FOREIGN KEY (following_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

