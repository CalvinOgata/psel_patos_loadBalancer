# PATOS PSEL 2.0 - Load Balancer in GO

Welcome to my submission for the PATOS Selection Process (PSEL 2.0). 

## About the Project

This repository contains a custom **Load Balancer built from scratch in Golang**. 

In accordance with the challenge guidelines, this project avoids high-level abstractions, web frameworks, and standard HTTP libraries (such as `net/http`). Instead, it handles raw network sockets (`net.Conn`), manages manual HTTP request/header parsing, and manages traffic distribution using low-level Go primitives.

## Features & Implementation Strategy

- **Raw Socket Operations:** Built directly on top of Go's `net` package for raw TCP listeners and connections.
- **Custom Request Parser:** Manual parsing of raw TCP streams into HTTP requests without external parsing libraries.
- **Load Balancing Strategy:** Implementation of custom balancing algorithms (such as Least Connections and Round Robin).
- **Active Health Checking:** Background routines monitoring backend server health via raw TCP probes.

---
*Documenting the journey, challenges, and architectural decisions made throughout the PSEL process.*
