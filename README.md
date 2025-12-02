## Accessing the API

1. **Clone the repository** 

2. **Create .env file:**
  ```
    DATABASE_URL="postgres://{$USERNAME}:{$PASSWORD}@localhost:5433/wallet_db?sslmode=disable"
    LOG_LEVEL=info
  ```

3. ### Start services: 
```
docker compose up -d
```

4. ### Stop services:
```
docker compose down
```
