# play-with-containers
This project aims to discover the container concepts and tools, and practice these tools by creating a microservices' architecture with docker and docker-compose.


inventory-app

istall postgresql:

sudo apt install postgresql postgresql-contrib
sudo -u postgres psql --version
sudo systemctl start postgresql
sudo systemctl enable postgresql
sudo -u pstgresql psql

-- Create the database
CREATE DATABASE movies_db;

-- Connect to the database
\c movies_db

-- Create the movies table
CREATE TABLE movies (
    id VARCHAR(255) PRIMARY KEY NOT NULL,
    title VARCHAR(255) NOT NULL,
    description VARCHAR(255)
);


TEST ENDPOINT:

1. Création d'un film
``` bash
    curl -X POST http://localhost:8080/api/movies \
    -H "Content-Type: application/json" \
    -d '{"id":"skjsjndhshbb123hjbhB32","title": "Inception", "description": "Un film sur les rêves"}'
```
2. Récupération de tous les films
``` bash
    curl http://localhost:8080/api/movies
```
3. Recherche de films par titre
``` bash
    curl "http://localhost:8080/api/movies?title=Inception"
```
4. Récupération d'un film spécifique
``` bash
    curl http://localhost:8080/api/movies/1
```
5. Mise à jour d'un film
``` bash
    curl -X PUT http://localhost:8080/api/movies/1 \
    -H "Content-Type: application/json" \
    -d '{"description": "Un film sur les rêves et la réalité"}'
```
6. Suppression d'un film
``` bash
    curl -X DELETE http://localhost:8080/api/movies/1
```
7. Suppression de tous les films (avec en-tête de confirmation)
``` bash
    curl -X DELETE http://localhost:8080/api/movies \
  -H "Confirm-Delete: yes"
```


First, clean up the existing containers:


docker compose down -v
docker rm -f inventory-database

For a complete fresh start (if needed):
docker compose down --volumes --rmi all
docker system prune -a


lien doc rabbitmq:
https://www.rabbitmq.com/docs/management-cli