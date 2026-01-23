-- Migración de prueba para test de deploy manual
-- Esta migración crea una tabla temporal para verificar el aislamiento de schemas
-- NO debe ejecutarse en producción

CREATE TABLE IF NOT EXISTS test_deploy_manual (
    id SERIAL PRIMARY KEY,
    test_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    test_data JSONB
);

-- Insertar un registro de prueba para verificar que funciona
INSERT INTO test_deploy_manual (test_name, test_data) 
VALUES ('test_deploy_manual', '{"branch": "test/deploy-manual-dev", "schema": "branch_test-deploy-manual-dev"}'::jsonb);

-- Comentario para identificar esta migración
COMMENT ON TABLE test_deploy_manual IS 'Tabla temporal de prueba para verificar deploy manual - NO usar en producción';
