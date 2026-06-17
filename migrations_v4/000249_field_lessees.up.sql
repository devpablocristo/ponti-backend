BEGIN;

-- Arrendatarios por campo (rol lessee). Tabla dedicada, espejo estructural de
-- field_investors, PERO apunta DIRECTO a actors(id): el arrendatario ES un actor con
-- rol lessee, no una fila de la tabla investors. Varios arrendatarios por campo, cada
-- uno con su porcentaje. Aditivo: no toca field_investors ni ninguna tabla existente.
CREATE TABLE public.field_lessees (
    field_id   bigint  NOT NULL,
    actor_id   bigint  NOT NULL,
    percentage integer NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone,
    created_by text,
    updated_by text,
    deleted_by text
);

ALTER TABLE ONLY public.field_lessees
    ADD CONSTRAINT pk_field_lessees PRIMARY KEY (field_id, actor_id);
ALTER TABLE ONLY public.field_lessees
    ADD CONSTRAINT fk_field_lessees_field FOREIGN KEY (field_id) REFERENCES public.fields(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.field_lessees
    ADD CONSTRAINT fk_field_lessees_actor FOREIGN KEY (actor_id) REFERENCES public.actors(id) ON DELETE CASCADE;

CREATE INDEX idx_field_lessees_actor_id ON public.field_lessees USING btree (actor_id);

COMMIT;
