<h1 align='center'>
  🌊 OpenAbyss 🌊
</h2>

<p align='center'>
 Secure encapsulated file storage solution. Interactions are done through Client-Server RPC, so the Client & Server can be anywhere!
</p>

## Configuration ⚙️
### Client
The following is the **default** generated Client configuration (*.config/config-client.json*).
```json
{
  "grpcName": "OpenAbyss-Client",
  "grpcHost": "localhost",
  "grpcPort": 50051,
  "insecure": false,
  "tlsCertPath": "cert/ca-cert.pem"
}
```
- `grpcName`: The name of the Client
- `grpcHost`: The host that the client grpc connects to
- `grpcPort`: The port that the client grpc connects to
- `insecure`: Secure by default. Inverse state of TLS. **Insecure=True** -> No TLS.
  - Overridden to **true** if **tlsCertPath** not found
- `tlsCertPath`: Path to the Client TLS Certificate


### Server
The following is the **default** generated Server configuration (*.config/config-server.json*).
```json
{
  "defaultKeyAlgorithm": "rsa",
  "insecure": false,
  "grpcPort": 50051,
  "grpcHost": "0.0.0.0",
  "tlsCertPath": "cert/server-cert.pem",
  "tlsKeyPath": "cert/server-key.pem",
  "backup": {
    "enable": true,
    "retentionPeriod": 604800000,
    "backupFrequency": 604800000
  }
}
```
- `defaultKeyAlgorithm`: Default algorithm used to generate keypair
- `insecure`: Secure by default. Inverse state of TLS. **Insecure=True** -> No TLS.
- `grpcPort`: The port that the server grpc listens to
- `grpcHost`: The host that the server grpc listens to
- `tlsCertPath`: Path to the Server TLS Certificate
- `tlsKeyPath`: Path to the Server TLS Key
- `backup`: Server backup settings
  - `enable`: Enabled state
  - `retentionPeriod`: Milliseconds to keep backup stored for
  - `backupFrequency`: Frequency in milliseconds to invoke backups


## TLS ⚙️
Server can be run without TLS, but if you'd like to generate a self-signed one to run **locally**,
```sh
_scripts/generate_certs.sh
```

❗ For **remote servers**, modify [server-ext.cnf](_scripts/server-ext.cnf) respectively.

## Installing OpenAbyss 📦
Binary home is installed under `/opt/OpenAbyss`. 
Binary is symlinked in `/usr/bin/open-abyss`

Run the following,
```sh
# Generate the certificates used between the client & server. Only have to
#  generate this once.
_scripts/generate_certs.sh

# Run the install script with sudo permissions.
_scripts/build.sh --install
```


## Build and Run 🚀
Building both `server` and `client` binaries by running the following,
```sh
# Builds both server and client binaries under "build" directory
_scripts/build.sh

# Running the server
./build/server

# Running the client
./build/client
```


## Client: Basic Usage 🤖
Commands are mostly client-side used to interact with the server.

At any time, passing in the `--help` flag, will print out the help menu with the client commands & argument usage.

### Generating/Listing Keys
```sh
# Generate a new keypair named "key1"
./build/client keys generate --name key1

# Listing stored keys
./build/client list keys
```

### Encrypting
```sh
# Encrypting a file called "file1", stores it in root server storage
./build/client encrypt --path ./file1 --key-id key1

# Encrypt file to "/some/path"
./build/client encrypt --path ./file1 --key-id key1 --storage-path=/some/path/
```

### Decrypting
```sh
# Decrypt file at "/file1" & output content to stdout
./build/client decrypt --path /file1 --key-id key1

# Decrypt file at "/some/path/file1" & output content to file.txt
./build/client decrypt --path /some/path/file1 --key-id key1 --out file.txt
```

### Listing Server Storage
```sh
# Listing server storage at root
./build/client list storage

# Listing server storage recursively
./build/client list storage --recursive

# Listing server storage recursively from "/some/path"
./build/client list storage --recursive --path /some/path
```



## License 📔
Licensed under the [MIT](LICENSE) License.
