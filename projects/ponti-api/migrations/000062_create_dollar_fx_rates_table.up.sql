-- =====================================================
-- 000062: DOLLAR - Tabla de Tipos de Cambio
-- =====================================================
-- Entidad: dollar (Dólares)
-- Funcionalidad: Almacenar tipos de cambio para conversiones USD/ARS
-- =====================================================

-- Crear tabla de tipos de cambio para conversiones monetarias
CREATE TABLE IF NOT EXISTS fx_rates (
    id SERIAL PRIMARY KEY,
    currency_pair VARCHAR(10) NOT NULL, -- Ej: USDARS, EURUSD
    rate DECIMAL(10,4) NOT NULL, -- Tasa de cambio
    effective_date DATE NOT NULL, -- Fecha de vigencia
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Crear índice único para evitar duplicados de tasa por fecha
CREATE UNIQUE INDEX IF NOT EXISTS idx_fx_rates_unique_pair_date 
ON fx_rates(currency_pair, effective_date);

-- Crear índice para búsquedas por fecha
CREATE INDEX IF NOT EXISTS idx_fx_rates_effective_date 
ON fx_rates(effective_date);

-- Insertar tasa inicial USD/ARS como placeholder
INSERT INTO fx_rates (currency_pair, rate, effective_date) 
VALUES ('USDARS', 1.0000, CURRENT_DATE)
ON CONFLICT (currency_pair, effective_date) DO NOTHING;

-- Comentarios en español
COMMENT ON TABLE fx_rates IS 'Tabla de tipos de cambio para conversiones monetarias';
COMMENT ON COLUMN fx_rates.currency_pair IS 'Par de monedas (ej: USDARS)';
COMMENT ON COLUMN fx_rates.rate IS 'Tasa de cambio';
COMMENT ON COLUMN fx_rates.effective_date IS 'Fecha de vigencia de la tasa';
