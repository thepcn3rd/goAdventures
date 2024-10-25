## Javascript - Get IP and User Agent
Note: Geolocation works with https unless disabled in the browser...

```javascript
<script>
	// Function to fetch the public IP address
	function getPublicIPAddress() {
  	return fetch('https://api64.ipify.org?format=json')
    		.then(response => response.json())
    		.then(data => data.ip)
    		.catch(error => {
      			console.error('Error fetching IP address:', error);
    		});
	}

	// Function to make a GET request using the IP address
	function makeGetRequest(ipAddress) {
	// Get the user agent string
	const userAgent = navigator.userAgent;
	const url = `http://54.218.226.238/?ip=${ipAddress}&ua=${userAgent}`; // Replace with your actual API endpoint

  	return fetch(url)
    		.then(response => response.json())
    		.then(data => {
      			console.log('Response from GET request:', data);
    		})
    		.catch(error => {
      			console.error('Error making GET request:', error);
    		});
	}

	// Usage
	getPublicIPAddress()
  		.then(ipAddress => {
    		makeGetRequest(ipAddress);
  	});
	
	</script>
```