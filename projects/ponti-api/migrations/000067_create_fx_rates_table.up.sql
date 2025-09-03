-- =====================================================
-- 000067: FX RATES - Tabla de Tipos de Cambio
-- =====================================================
-- Entidad: FX Rates (Tipos de Cambio)
-- Funcionalidad: Almacenar tasas de cambio para conversiones de moneda
-- =====================================================

-- Crear tabla de tipos de cambio si no existe
CREATE TABLE IF NOT EXISTS fx_rates (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(10) NOT NULL, -- Ej: USDARS, EURUSD, etc.
    rate NUMERIC(18,6) NOT NULL, -- Tasa de cambio
    as_of_date DATE NOT NULL, -- Fecha de la tasa
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    created_by BIGINT,
    updated_by BIGINT,
    deleted_by BIGINT,
    
    -- Índices para optimizar consultas
    CONSTRAINT uk_fx_rates_code_date UNIQUE (code, as_of_date)
);

-- Crear índices para optimizar consultas
CREATE INDEX IF NOT EXISTS idx_fx_rates_code ON fx_rates(code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_fx_rates_as_of_date ON fx_rates(as_of_date) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_fx_rates_deleted_at ON fx_rates(deleted_at);

-- Insertar tasa inicial para USDARS si la tabla está vacía
-- Comentado: INSERT INTO fx_rates (code, rate, as_of_date, created_at, updated_at)
-- VALUES ('USDARS', 1.0, CURRENT_DATE, now(), now());
