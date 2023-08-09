# Url-Shortener

A url shortening service made using Gofiber. Redis is used as the database to store the shortened url, custom short url and the expiry time. Api quota limit is set to be 10 as of now. uuid package is used to get a custom short url and govalidator is used to check if the given json input is url or not. To start with the project simply run this command: docker compose up -d, and use thunder client, Postman, or any other platform to interact with the API.
