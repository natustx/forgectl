.PHONY: build clean

build:
	cd forgectl && go build -o forgectl .

clean:
	rm -f forgectl/forgectl
