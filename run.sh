if [ -f .env ]
then
    export $(cat .env | xargs)
fi
go run .
