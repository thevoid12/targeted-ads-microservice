# high level todo list
- [x] learn and understand gokit documentation
- [x] design
- [x] choose db
- [x] choose a cache
- [x] insert random data
- [x] implement core go logic
- [x] implement updating cache when db changes
- [ ] write unit test
- [ ] add grafana 
- [x] update readme 

# micro level todos
- [ ] remove errorf statement and add logger for pgsql db
- [x] we need 2 tables 1. targeting rules,2.campaign details targetting rules will have campign id as foreign key
- [x] select all data from postgres and create our inmemory inverted index cache on restarts
- [x] write the main logic using gokit format 
- [x] redis stream setup
- [x] notfication when pg db is updated
- [ ] tests for all of it
- [x] updating the cache without reloading based on redis stream
- [x] send the result back to the client after the fetch from the cache
- [x] change context.TODO to appropriate context