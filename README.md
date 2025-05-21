# Ponti API

1. Ejecutar: make stg-build

2. Crear 2 crops, ejemplo:

curl --location 'localhost:8080/api/v1/crops/public/' \
--header 'Content-Type: application/json' \
--data '{
"name":"wheat"
}'

3. Probar los endpoints de la coleccion Soalen.

INSERT INTO crops (id, name) VALUES
(1, 'Soja'),
(2, 'Maíz'),
(3, 'Trigo'),
(4, 'Girasol'),
(5, 'Sorgo'),
(6, 'Cebada'),
(7, 'Alfalfa'),
(8, 'Maní'),
(9, 'Centeno'),
(10, 'Avena');
