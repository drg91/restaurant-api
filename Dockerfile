
# Etapa de compilación
FROM golang:1.21.1-alpine AS builder

# Establece el directorio de trabajo
WORKDIR /app

# Copia los archivos del módulo y descarga las dependencias
COPY go.mod ./
RUN go mod download

# Copia el código fuente
COPY . .

# Compila el código
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o restaurant-api .

RUN apk add --no-cache tzdata
ENV TZ=America/Argentina/Cordoba

# Etapa de ejecución
FROM alpine:latest

WORKDIR /root/

# Copia el ejecutable
COPY --from=builder /app/restaurant-api .

# Puerto de exposición
EXPOSE 8080

# Comando para ejecutar la aplicación
CMD ["./restaurant-api"]
