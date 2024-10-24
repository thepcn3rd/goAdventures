# Compiled Javascript v8

#### Summary

While reading about threat actors using compiled JavaScript with the V8 engine, I decided to experiment with it myself. I found a library called v8go, which I used to serve compiled JavaScript along with HTML on a web server. Additionally, I learned how to install NodeJS and write code that interacts with the Windows API. I also figured out how to compile this code into an executable file for Windows, and the equivalent formats for Linux and Mac.

#### Creating a Compiled Javascript v8 File

createCompiledJSC.go - This program at current allows you to place javascript in it and it compiles it into a v8 jsc file.  To enhance this I would create a flag that would read the javascript from a given file.

main.go - This program takes the javascript, compiles it once and then serves it on a web site.  However it is not working the way that I expect.  I was hoping to load the pre-compiled JSC file and only run it.  The initial execution does compile the javascript then it uses the cached data.  I thought I would place this project on hold because I was not getting very far with it, it also appears that it is not possible what I was doing, however it seems like it should be.

![Wolf Howling at the Moon in Metallic Armor](/projects/compileJavascriptV8/wolfMoon.jpg)

#### Node JS - Installing and Configuring

Utilizing Manjaro as my host OS I installed NodeJS and npm.  npm is the node package manager.

```bash
sudo pacman -S nodejs npm
```

I created a nodejs folder and then a helloWorld folder inside of that.  To initialize the Node project in that helloWorld directory the following command needed to be executed.

```bash
npm init -y
npm install --save express
```

Created some simple code for NodeJS to create a web server on port 3000 to display the words Hello World as it is visited.  

```javascript
const http = require('http');

const hostname = '127.0.0.1';
const port = 3000;

const server = http.createServer((req, res) => {
    res.statusCode = 200;
    res.setHeader('Content-Type', 'text/html');
    res.end('<html><body><h1>Hello, World!</h1></body></html>');
});

server.listen(port, hostname, () => {
    console.log(`Server running at http://${hostname}:${port}/`);
});
```

To execute the code and serve it with node the following command was executed.
```bash
node server.js
```

Then I discovered I needed a package called bytenode to be able to use it to compile the code.  To install this I had to run a couple of different commands and I may change this as I test it out later
```bash
sudo npm install --save -g bytenode
npm install --save -g bytenode
```

After installing the bytenode package I modified the below code to create the javascript compiled code and save it as JSC.
```javascript
const bytenode = require('bytenode');

// Compiling JavaScript into bytecode and executing it
bytenode.compileFile('server.js', 'server.jsc'); // Compiling JavaScript to bytecode
```

Used Node to Execute the Code
```bash
node compile.js
```

Then to run the compiled JSC file the following code was created.
```javascript
require('./server.jsc'); // Running the compiled code
```

Used Node to Execute the Code
```bash
node runCompiled.js
```

Then to package it for Windows, Mac OS and Linux first I had to install pkg.
```bash
sudo npm install --save -g pkg
npm install --save -g pkg
```

Packaging the javascript required minimal changes to the package.json file and then executed the following
```bash
pkg . --targets node14-win-x64,node14-mac-x64,node14-linux-x64 --without-intl
```

Here is the package.json file that I used, I am not sure all of the modifications that I made
```json
{
  "name": "helloworld",
  "version": "1.0.0",
  "main": "index.js",
  "scripts": {
    "test": "echo \"Error: no test specified\" && exit 1"
  },
  "keywords": [],
  "author": "",
  "license": "ISC",
  "description": "",
  "bin": "server.js",
  "dependencies": {
    "bytenode": "^1.5.6",
    "express": "^4.19.2"
  },
  "pkg": {
    "assets": [],
    "scripts": []
  }
}
```

#### References

https://research.checkpoint.com/2024/exploring-compiled-v8-javascript-usage-in-malware/

https://pkg.go.dev/rogchap.com/v8go

https://github.com/rogchap/v8go


  






