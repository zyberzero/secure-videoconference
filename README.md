# HoldSpace

Holdspace is a video conference system where the user is identified by Bank ID.


It is based on [ION](https://www.github.com/pion/ion/). 

## Prerequisites

* GrandID credentials, A BankID integration solution from [E-identitet AB](https://www.e-identitet.se).
* Docker and docker-compose

## How to use

### 1. Set env

#### Mandatory environment variables

* `GRANDID_API`
* `GRANDID_SERVICE`

#### Optional environment variables

`MDB_SQLITE_KEY` for overriding the default encryption key for the SQLite database used by the MDB (Meeting database) service.

Optional (if provided, the server needs to be accessible by the domain given in order to generate certificates from LetsEcrypt):

```
export WWW_URL=yourdomain
export ADMIN_EMAIL=yourname@yourdomain
```

### 2. Deployment
#### 1. clone
```
git clone https://github.com/zyberzero/secure-videoconference/
```

#### 2. run
```
docker-compose up
```

#### 3. chat
If `WWW_URL` is not set, open `http://localhost:8080`, otherwise `https://$WWW_URL:8080` in a Chrome browser.

