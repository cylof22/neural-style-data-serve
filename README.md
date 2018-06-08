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

		go get github.com/rs/cors
		
          If no VPN, please create the folder src/golang/org/x and git clone https://github.com/golang/image .
		
	 (2) Define the local GOPATH environment for the go build
	 
	    export GOPATH="source-code-folder"
			
	 (3) Build the Product, Web and User Service
	 
	    cd source-code-folder/src/neural-style-data-server
	    
	    go install -v ./...
	    
	 (4) Build the Azure Cloud Storage Service
	 
	    cd source-code-folder/src/neural-style-image-store
	    
	    go install -v ./...
	    
			 
	 (5) Now only two Services. The Products, and User Service, and User Service should be independent services in 	   
	     future.
	 
	 (6) MongoDB 
	 
	     Launch the MongoDB: mongod --port 9000
	 
	 (7) Memcached
	     
	     Launch the Memcached: use the default port now. Need more discussions for the memcached service deployment.
	     
2.  Deploy Process
    
	 (1) Web Service, Products, and User Service
	 
	     The Basic command arguments are: 
	     host          = Service Address, default is 0.0.0.0
	     port          = Service Port: default is 8000
	     dbserver      = MongoDB Server Address: default 0.0.0.0, need login information in future.
	     dbport        = MongoDB Server Port: default is 9000
	     storageURL    = Azure Cloud Storage Service Address: default is 0.0.0.0
	     storagePort   = Azure Cloud Storage Service Port: default is 5000
	     cacheHost     = Memcached Service Address: default is www.elforce.net. Need to access from the web, so use the 
	         	     External products address. In future the value will be a a group of address:port which is 
			     seperated by ';', This is done by the product service, and is not related with the deployment.
			     
	     The Basic Environments are 
	     TOKEN_KEY: used by the user service to parse the jwt token.
	 (2) Azure Cloud Storage Service
	 
	     The Basic command arguments are:
	     host         = Service address, default is 0.0.0.0
	     port         = Service Port: default is 5000
	     dbserver     = MongoDB Service Address: default is 0.0.0.0. Need login information in future.
	     dbport       = MongoDB Service Port: default is 9000.
	     
	     The Basic Enviroments: 
	     MAX_WORKERS           = Internal Storage Engine worker size, default value is 2 now.
	     MAX_QUEUE             = Internal Storage Engine job queue size, default value is 2 now.
	     AZURE_STORAGE_ACCOUNT = Azure Storage Account, only one string. In future, it will be a group of storage  					     accounts seperated by ';'. 
	     AZURE_STORAGE_KEY     = Azure Storage Account key, only one string now. Like account string, it will be a group  					   of key.
	     AZURE_STORAGE_URL     = Azure Storage URL. For china, '.blob.core.chinacloudapi.cn', and for others, 					     '.blob.core.windows.net'.
	
	
