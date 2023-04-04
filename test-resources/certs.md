## Generating keys with given rootCA.pem and rootCA.key.pem
### Cert1 
openssl genrsa -des3 -out server1.key 4096
openssl pkcs8 -topk8 -inform PEM -outform PEM -nocrypt -in server1.key -out key.pem
rm server1.key

CSR 
openssl req -new -key key.pem -out server1.csr

openssl x509 -req -days 36500 -in server1.csr -CA rootCA.pem -CAkey rootCA.key.pem -CAcreateserial -out cert.pem
rm server1.csr

### Cert2

openssl genrsa -des3 -out server2.key 4096
openssl pkcs8 -topk8 -inform PEM -outform PEM -nocrypt -in server2.key -out key2.pem
rm server2.key

CSR 
openssl req -new -key key2.pem -out server2.csr

openssl x509 -req -days 36500 -in server2.csr -CA rootCA.pem -CAkey rootCA.key.pem -CAcreateserial -out cert2.pem
rm server2.csr
