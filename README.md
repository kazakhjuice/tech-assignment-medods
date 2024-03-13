## to run mongo:

docker run --name some-mongo -p 27017:27017 -d mongo

## to run server

go run cmd/main.go

# two endpoints:

## /login - get 

recieves json in format

{
  "uuid":"something"
}

## /update - patch

recieves json in format 

{
    "jwt": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MTAzNTM1MjMsImp0aSI6ImFkaWwifQ.7Pgv34gcuIXbaJpT6KZ0TvJ8Jfm8y8DmsE27eXwoTsggbJp3NhRxtHO32XQvXyJV6sceIAHq0BruCsaWamjMpw",
    "refreshToken": "EbuwMlL32JNV9usFCKw3Xa8JXMz3eR0e5fWgQKRsES8="
}

