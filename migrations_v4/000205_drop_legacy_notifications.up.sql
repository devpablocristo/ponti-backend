-- Drop del sistema legacy de notifications (ponti_notifications + ponti_notification_candidates).
-- Reemplazado por businessinsights (migracion 204) con patron event-driven estilo pymes.
-- Las tablas se crearon en la migracion 203 y ya no tienen consumidores en el codigo.

BEGIN;

DROP TABLE IF EXISTS public.ponti_notifications CASCADE;
DROP TABLE IF EXISTS public.ponti_notification_candidates CASCADE;

COMMIT;
