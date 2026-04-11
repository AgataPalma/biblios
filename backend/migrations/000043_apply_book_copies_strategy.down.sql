DROP TRIGGER IF EXISTS user_copy_states_updated_at ON public.user_copy_states;

DROP INDEX IF EXISTS public.idx_user_copy_states_user_status;
DROP INDEX IF EXISTS public.idx_user_copy_states_copy;
DROP INDEX IF EXISTS public.idx_user_copy_states_user;

DROP TABLE IF EXISTS public.user_copy_states;
