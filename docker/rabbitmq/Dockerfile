FROM debian:bookworm

ENV DEBIAN_FRONTEND=noninteractive
ENV RABBITMQ_VHOST=/

# Copier et exécuter le script d'installation
COPY rabbitmq.sh /tmp/rabbitmq.sh
RUN chmod +x /tmp/rabbitmq.sh && \
    /tmp/rabbitmq.sh && \
    rm /tmp/rabbitmq.sh && \
    rm -rf /var/lib/apt/lists/*

# Exposer les ports
EXPOSE 5672 15672

# Créer le répertoire de logs
RUN mkdir -p /var/log/rabbitmq && \
    chown rabbitmq:rabbitmq /var/log/rabbitmq

# Script de démarrage
COPY start_rabbitmq.sh /start_rabbitmq.sh
RUN chmod +x /start_rabbitmq.sh

# Changer vers l'utilisateur non-root
USER rabbitmq

# Commande par défaut
CMD ["/start_rabbitmq.sh"]