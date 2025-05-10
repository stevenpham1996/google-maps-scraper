# Commands:

1. Set up the PostgreSQL database:

```bash

# Create the database and user

sudo -u postgres psql -c "CREATE DATABASE gmapsdb;"
sudo -u postgres psql -c "CREATE USER gmapsuser WITH PASSWORD 'justbeginagain';"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE gmapsdb TO gmapsuser;"


# Run the setup script

psql -h localhost -p 5432 -U gmapsuser -d gmapsdb -W -f setup_database.sql

```

2. Start docker container:`docker-compose -f docker-compose.postgres.yaml up -d`

3. Start App server:`./google-maps-scraper -dsn ostgres://gmapsuser:justbeginagain@localhost:5432/gmapsdb" -c 16 -data-folder ./gmapsdata`
