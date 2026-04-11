-- Apply single-owner inventory + per-user reading state strategy

CREATE TABLE IF NOT EXISTS public.user_copy_states (
                                                       user_id UUID NOT NULL,
                                                       copy_id UUID NOT NULL,
                                                       reading_status VARCHAR(20) NOT NULL DEFAULT 'want_to_read',
                                                       current_page INTEGER,
                                                       started_reading_at TIMESTAMPTZ,
                                                       finished_reading_at TIMESTAMPTZ,
                                                       reread_count INTEGER NOT NULL DEFAULT 0,
                                                       personal_notes TEXT,
                                                       created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                                                       updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                                                       CONSTRAINT user_copy_states_pkey PRIMARY KEY (user_id, copy_id),
                                                       CONSTRAINT user_copy_states_reading_status_check CHECK (
                                                           reading_status IN ('want_to_read', 'reading', 'read', 'did_not_finish')
                                                           ),
                                                       CONSTRAINT user_copy_states_current_page_check CHECK (
                                                           current_page IS NULL OR current_page >= 0
                                                           ),
                                                       CONSTRAINT user_copy_states_reread_count_check CHECK (reread_count >= 0),
                                                       CONSTRAINT user_copy_states_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE,
                                                       CONSTRAINT user_copy_states_copy_id_fkey FOREIGN KEY (copy_id) REFERENCES public.book_copies(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_user_copy_states_user
    ON public.user_copy_states(user_id);

CREATE INDEX IF NOT EXISTS idx_user_copy_states_copy
    ON public.user_copy_states(copy_id);

CREATE INDEX IF NOT EXISTS idx_user_copy_states_user_status
    ON public.user_copy_states(user_id, reading_status);

CREATE TRIGGER user_copy_states_updated_at
    BEFORE UPDATE ON public.user_copy_states
    FOR EACH ROW
EXECUTE FUNCTION public.update_updated_at();

-- Backfill the owner's reading state from existing columns in book_copies
INSERT INTO public.user_copy_states (
    user_id,
    copy_id,
    reading_status,
    current_page,
    started_reading_at,
    finished_reading_at,
    reread_count,
    personal_notes,
    created_at,
    updated_at
)
SELECT
    bc.owner_id,
    bc.id,
    bc.reading_status,
    bc.current_page,
    bc.started_reading_at,
    bc.finished_reading_at,
    bc.reread_count,
    bc.personal_notes,
    bc.created_at,
    bc.updated_at
FROM public.book_copies bc
ON CONFLICT (user_id, copy_id) DO NOTHING;
