# Toy banklink
A Toy banklink "clone" built for educational purposes

## Disclaimer
Mooncascade is not affiliated with or endorsed by Truelayer and this repository is published for educational purposes only

## Setting up local development environment
Configurations and tasks needed to set up local development environment.

### Prerequisites
- Go
- Docker

### Configuration
The configuration file must exist at the following path: 
```
'./server/infra/config/truelayer/config.json'
```

The configuration template can be found from:
```
'./server/infra/config/truelayer/config.json.example'
```

### Running locally
Create go vendor folder.
```sh
cd server
go mod vendor
```
Bring up both the API and UI with the included composefile
```
docker-compose up
```

### Documentation
After the project has started, documentation can be read [here](http://localhost/swagger.html)

## Quick start
Instructions on how to go through the main flow:

1. Prepare payment with '/payment' endpoint and receive the UUID (example curl below)
    
    ```
        curl --location --request POST 'http://localhost:3000/api/payment' \
        --header 'Content-Type: application/json' \
        --data-raw '{
            "receiver_id": "1234",
            "amount": 100
        }'
    ```
2. Navigate to 'http://localhost/index.html?uuid={received_uuid}' with browser
3. Select bank and browser will get redirected to the bank payment site.
4. Proceed with regular bank payment flow
    - Demo bank credentials:
        - Customer number: `123456789012`
        - PIN: `572`
        - password: `436`
5. After completing or cancelling the payment, browser redirects back to initial payment screen and shows the new status.

