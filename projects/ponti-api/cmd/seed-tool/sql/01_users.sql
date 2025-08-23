-- Crear usuarios de prueba para el sistema
-- Los usuarios se crean para pruebas y desarrollo del sistema
-- Incluye usuarios con diferentes roles y estados de verificación para testing

INSERT INTO users (id, email, username, password, token_hash, refresh_tokens, id_rol, is_verified, active, created_by, updated_by, created_at, updated_at)
VALUES 
  (1, 'seed@local', 'seed', 'seedpwd', 'seedtoken', '{}', 1, TRUE, TRUE, 2, 2, NOW(), NOW()),
  (2, 'seed123@local', 'seed123', 'seedpwd', 'seedtoken', '{}', 1, TRUE, TRUE, 2, 2, NOW(), NOW()),
  (123, 'user123@example.com', 'user123', 'password123', 'token123', '{}', 1, TRUE, TRUE, 1, 1, NOW(), NOW());

-- Verificar inserción
SELECT '✅ Usuarios creados:' as status, COUNT(*) as total FROM users WHERE email LIKE 'seed%@local' OR id = 123; 