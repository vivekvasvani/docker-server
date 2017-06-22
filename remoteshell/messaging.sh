./remoteshell -c messaging.yml  -i ~/.ssh/id_rsa -s prod -cmd create -y
./remoteshell -c messaging.yml  -i ~/.ssh/id_rsa -s prod -cmd clone -y -async
./remoteshell -c messaging.yml  -i ~/.ssh/id_rsa -s prod -cmd start -y
