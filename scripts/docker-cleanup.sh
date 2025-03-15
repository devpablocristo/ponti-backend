#!/bin/bash

# Solicitar confirmación del usuario
read -p "Para continuar con la limpieza completa de Docker, escribe 'clean up' y presiona Enter: " confirm

# Verificar la entrada del usuario
if [ "$confirm" != "clean up" ]; then
  echo "Acción cancelada. No se realizó ninguna limpieza."
  exit 0
fi

# Detener y eliminar todos los contenedores
if [ "$(docker ps -aq)" ]; then
  docker stop "$(docker ps -aq)"
  docker rm "$(docker ps -aq)"
else
  echo "No hay contenedores para detener o eliminar."
fi

# Eliminar todas las imágenes
if [ "$(docker images -q)" ]; then
  docker rmi "$(docker images -q)"
else
  echo "No hay imágenes para eliminar."
fi

# Eliminar todos los volúmenes
if [ "$(docker volume ls -q)" ]; then
  docker volume rm "$(docker volume ls -q)"
else
  echo "No hay volúmenes para eliminar."
fi

# Eliminar todas las redes
if [ "$(docker network ls -q)" ]; then
  docker network rm "$(docker network ls -q)"
else
  echo "No hay redes para eliminar."
fi

# Limpiar el sistema de Docker
docker system prune -a --volumes -f

echo "Docker cleanup completed."
