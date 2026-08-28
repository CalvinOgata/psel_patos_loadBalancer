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

### Day 0: Getting Started

Laid the groundwork for the project today. I updated `README.md` with the overall project scope, set up `main.go` for the initial core logic, and built a basic `index.html` interface to make testing visual rather than terminal-bound. 

Most of my time was spent studying TCP/IP and HTTP fundamentals—specifically how HTTP is essentially structured text parsing over a network stream. I also spent time dissecting Go's standard libraries (`net`, `bufio`) to understand precisely how each function and connection operates under the hood.

### Day 1: System Architecture & Raw Backend Prototype

Took a step back today to refine the overall system architecture. I mapped out the core interaction between three primary components: the frontend interface, the load balancer, and the backend server. To keep the codebase clean and follow idiomatic Go project conventions, I reorganized the load balancer and backend entry points into separate subdirectories (`balancer` and `server`).

I also evaluated whether to run backend instances inside Docker containers or as standalone Go processes. I decided to start with native Go processes to maintain direct visibility into OS socket behavior before introducing containerization overhead. Finally, I wrote an initial prototype for the raw TCP backend server—it's still bare-bones, but it lays the foundation for image serving in the coming days.
