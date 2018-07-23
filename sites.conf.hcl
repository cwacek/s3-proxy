
site {
  Host = "foobar.local:8080"
  AWSKey = "AKIAIJBKYV4IOLHMQZDQ"
  AWSSecret = "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"
  AWSRegion = "us-east-1"
  AWSBucket = "test.projectblink.com"
  
  options {
    website = true
    prefix = "/docs2"
    }
  }

site {
  Host = "localhost:8080"
  AWSKey = "AKIAIJBKYV4IOLHMQZDQ"
  AWSSecret = "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"
  AWSRegion = "us-east-1"
  AWSBucket = "test.projectblink.com"
  
  options {
    website = true
    prefix = "/docs"
    }
  }
