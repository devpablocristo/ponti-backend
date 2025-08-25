-- =======================
-- USUARIO DEMO PARA LIST LOT
-- =======================
-- Usuario de prueba para generar datos de list lot

INSERT INTO users (id, email, username, password, token_hash, refresh_tokens, id_rol, is_verified, active, created_by, updated_by, created_at, updated_at)
VALUES (
    123, 
    'demo@ponti.com', 
    'demo_user', 
    'demo_password', 
    'demo_token_hash', 
    '{}', 
    1, 
    TRUE, 
    TRUE, 
    1, 
    1, 
    NOW(), 
    NOW()
);

-- Verificar inserción
SELECT '✅ Usuario demo creado' as status, id, email, username FROM users WHERE id = 123;
