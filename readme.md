

curl -X POST \
     -H "Content-Type: application/json" \
     -d '{
         "app": "test_app_id",
         "os": "android",
         "country": "US"
     }' \
     http://localhost:9090/v1/delivery