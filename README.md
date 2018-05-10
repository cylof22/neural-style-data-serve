# neural-style-data-serve

1. Build Process

   (1) Install dependent libraries
   		cd source-code-folder
			
   		go get github.com/go-kit/kit
			
		go get github.com/gorilla/mux
			
		go get github.com/go-logfmt/logfmt
			
		go get github.com/go-stack/stack
		
		go get gopkg.in/mgo.v2
		
		go get github.com/dgrijalva/jwt-go
		
		go get github.com/Azure/azure-pipeline-go
		
		go get -u github.com/aws/aws-sdk-go
		
		go get github.com/Azure/azure-storage-blob-go
		
		go get github.com/bradfitz/gomemcache/memcache
		
		go get golang.org/x/image
		
	 (2) Define the local GOPATH environment for the go build
	 
	    export GOPATH="source-code-folder"
			
	 (3) Build the transfer server
	 
	    cd source-code-folder/src/neural-style-data-server
			 
		go install -v ./...
			 
   (4) After executing the go install command, the source-code-folder/bin will be created. You can see the neural-style-data-server
		   program.
			 
	 (5) Build docker image. When launching the docker, the shared volumes for the network and image folder is required.
	 
	    cd source-code-folder
			 
		docker build -t neural-style-data-server .
		
			 
2. Launch transfer serve
   The server has three configuration parameters: "host", "port", and "network". The "host" and "port" are the server's network address.
	 The default values is "localhost:9090". The default value of the network is empty, which means that the "imagenet-vgg-verydeep-19.mat" 
	 is placed in the same folder with the "neural-style-data-server" program.
   
	 The HTTP server for the style transfer is GET "server address"//styleTransfer/?content="content image location"&style="style image location"
	 &output="output image location"&iterations = 100.
	 
	 The content, style, and output image location is encrypted by using base64.StdEncoding. You can see it in the transport.go.
	 
	
	
