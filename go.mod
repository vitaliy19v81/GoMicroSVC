module go_microsvc

go 1.21

toolchain go1.23.2

require (
	github.com/go-chi/chi/v5 v5.1.0
	github.com/joho/godotenv v1.5.1
	github.com/segmentio/kafka-go v0.4.47
	gorm.io/driver/postgres v1.5.9
	gorm.io/gorm v1.25.12
)

require (
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.1 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/crypto v0.27.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/text v0.18.0 // indirect
)

//replace github.com/go-chi/chi/v5 => /home/vtoroy/GolandProjects/tmp/chi/chi-5.1.0

//replace github.com/jackc/pgx/v5 => /home/vtoroy/GolandProjects/tmp/pgx/pgx-5.7.1

//replace github.com/joho/godotenv => /home/vtoroy/GolandProjects/tmp/godotenv/godotenv-1.5.1

//replace github.com/klauspost/compress => /home/vtoroy/GolandProjects/tmp/compress

//replace github.com/segmentio/kafka-go => /home/vtoroy/GolandProjects/tmp/kafka/kafka-go-0.4.47

//replace gorm.io/driver/postgres => /home/vtoroy/GolandProjects/tmp/postgres/gorm.io/driver/postgres

//replace gorm.io/gorm => /home/vtoroy/GolandProjects/tmp/gorm/gorm.io/gorm
