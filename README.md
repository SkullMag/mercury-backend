# Backend for Mercury - English Dictionary
This repository contains code for backend of my english dictionary web application

## Why "Mercury"? 
In my trip to the USA I have bought the book "Cosmic Queries" written by Neil deGrasse Tyson. However, my mother tongue is Russian, 
so I have to look up some words. This book has liven my curiosity, so I decided to name the application "Mercury" as Mercury is a nearest planet to the Sun
and for me it represents curiosity of mankind.

## Project structure
[Handlers](handlers/) contains of files where API handlers are implemented. They are splitted into different files for 
better navigation and reusability.
[Models](models/) contains of models for GORM. In production I currently use PostgreSQL which is defined in [this file](database/db.go) 
but I also used SQLite on start.
[Utils](utils/utils.go) is the place, where all reusable functions are implemented.


## Collaborate
Feel free to collaborate and make pull requests 😁 
