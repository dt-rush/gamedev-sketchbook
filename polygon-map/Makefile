all: main cpu.prof

test:
	go test -bench=. -cpuprofile

clean:
	rm main *.prof

main: 
	go build -o main *.go 

cpu.prof: main
	./main -cpuprofile=cpu.prof
