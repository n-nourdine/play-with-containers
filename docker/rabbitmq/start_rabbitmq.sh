#!/bin/bash
set -e

# Variables d'environnement avec valeurs par défaut
RABBITMQ_USER=${RABBITMQ_USER:-nasdev}
RABBITMQ_PASSWORD=${RABBITMQ_PASSWORD:-passer}
RABBITMQ_VHOST=${RABBITMQ_VHOST:-/}
RABBITMQ_QUEUE=${RABBITMQ_QUEUE:-billing_queue}

echo "🐰 Démarrage de RabbitMQ..."
echo "Utilisateur: ${RABBITMQ_USER}"
echo "VHost: ${RABBITMQ_VHOST}"
echo "Queue: ${RABBITMQ_QUEUE}"

# Démarrer RabbitMQ en arrière-plan
rabbitmq-server &
PID=$!

# Attendre que RabbitMQ soit prêt
echo "⏳ Attente du démarrage de RabbitMQ..."
until rabbitmqctl ping > /dev/null 2>&1; do
    echo "Attente de RabbitMQ..."
    sleep 2
done
echo "✅ RabbitMQ est prêt!"

# Activer le plugin de management
echo "🔧 Activation du plugin de management..."
rabbitmq-plugins enable rabbitmq_management

# Attendre que le plugin soit prêt
echo "⏳ Attente du plugin de management..."
sleep 5

# Créer un utilisateur, un tag et des permissions
echo "👤 Création de l'utilisateur ${RABBITMQ_USER}..."
rabbitmqctl add_user ${RABBITMQ_USER} ${RABBITMQ_PASSWORD} || true
rabbitmqctl set_user_tags ${RABBITMQ_USER} administrator || true
rabbitmqctl set_permissions -p ${RABBITMQ_VHOST} ${RABBITMQ_USER} ".*" ".*" ".*" || true

# Déclarer une queue
echo "📋 Création de la queue ${RABBITMQ_QUEUE:-billing_queue}..."
rabbitmqadmin declare queue name=billing_queue durable=true

echo ""
echo "🎉 RabbitMQ configuré avec succès!"
echo "🌐 Interface: http://localhost:15672"
echo "👤 Utilisateur: ${RABBITMQ_USER} / ${RABBITMQ_PASSWORD}"
echo "📋 Queue: ${RABBITMQ_QUEUE:-billing_queue}"

# Garder le conteneur actif
wait $PID