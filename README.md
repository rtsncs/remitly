# Remitly Internship Exercise

## Running
### With Docker

1. Clone the repository:

```bash
git clone https://github.com/rtsncs/remitly.git
cd remitly
```

2. Create a `.env` file:

```bash
DATABASE_USERNAME=postgres
DATABASE_PASSWORD=postgres
DATABASE_NAME=swift
API_PORT=8080
```

3. Start services:

```bash
docker compose up --build -d
```

4. Load initial data from a spreadsheet:

```bash
docker compose cp /path/to/spreadsheet.xlsx api:/data.xlsx
docker compose exec api "./app load -f /data.xlsx"
```

5. The API will be accessible at `localhost:8080`:
```bash
curl http://localhost:8080/v1/swift-codes/PTFIPLPWAAP
```
```json
{"address":"UL CHLODNA 52  WARSZAWA, MAZOWIECKIE, 00-872","bankName":"PKO TOWARZYSTWO FUNDUSZY INWESTYCYJNYCH SA","countryISO2":"PL","countryName":"POLAND","isHeadquarter":false,"swiftCode":"PTFIPLPWAAP"}
```

### Without Docker
Prerequisites:
- [Go](https://go.dev/doc/install)
- [PostgreSQL](https://www.postgresql.org/download/)

1. Clone the repository:

```bash
git clone https://github.com/rtsncs/remitly.git
cd remitly
```

2. Export database URL (assuming PostgreSQL is running and database `swift` is already created):
```bash
export DATABASE_URL=postgresql://localhost/swift
```

3. Load initial data from a spreadsheet:
```bash
go run main.go load -file=/path/to/spreadsheet.xlsx
```

4. Run the server:
```bash
go run main.go serve
```

5. The API will be accessible at `localhost:8080`:
```bash
curl http://localhost:8080/v1/swift-codes/PTFIPLPWAAP
```
```json
{"address":"UL CHLODNA 52  WARSZAWA, MAZOWIECKIE, 00-872","bankName":"PKO TOWARZYSTWO FUNDUSZY INWESTYCYJNYCH SA","countryISO2":"PL","countryName":"POLAND","isHeadquarter":false,"swiftCode":"PTFIPLPWAAP"}
```

## Testing
Prerequisites:
- [Go](https://go.dev/doc/install)
- [Docker Compose](https://docs.docker.com/compose/install/)

To run the tests run:
```bash
go test -v ./...
```
