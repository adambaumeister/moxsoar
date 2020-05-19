# MOXSOAR 
## Description
A mocking utility designed to replicate commonly integrated system APIs for use in testing automation pipelines.

## Get started

```bash
docker run --name moxsoar -d -p 8000-8100:8000-8100
```
After starting, the moxsoar UI is available at http://[your-server]:8080. Login with **admin/admin**.

## Docker options
**ports**

MOXSOAR uses multiple ports to accurately mock various HTTP API implementations on the same IP address.

By default these ports are in the range 8000-8100, so you should always use 

```bash
-p 8000-8100:8000-8100
``` 

**Persistent Volumes**

*/etc/moxsoar/data*: Stores persistent data such as the user database

**ephemeral volumes**

*/etc/moxsoar/content* : Stores all Mock content - useful to mount for making changes on the fly

*/etc/moxsoar* : The primary moxsoar directory - contains the intial config file, and in practice, you will rarely need to edit this.

