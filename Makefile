build: create_output_dir
	go build -pkgdir output/pkg/ -i -o output/ -v -trimpath ./...

alpine: create_output_dir
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo  -pkgdir output/pkg/ -i -o output/ -v -trimpath ./...


create_output_dir:
	mkdir -p output/ output/pkg/

clean:
	rm -rf output/
