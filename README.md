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

### Day 2: Basic Proxy, Run Script & Working Image Delivery

Today was a tough one, but it got the core of the project working end to end. On the backend I tightened up the error handling, restricted the served files to PNG and JPEG, and added an `images` folder so the application finally has something real to hand out. I also reworked the frontend (`index.html`) so it stops announcing "Hello World" and actually displays the images it receives. To avoid juggling two terminals every time I wanted to test something, I wrote a small `run.sh` that starts the backend server and the load balancer together.

The backend was by far the most time-consuming part. Adding new routes and reshaping them as the application kept evolving was challenging, and getting everything to line up—the manual HTTP parsing, the ports, and Go's syntax, which is fun and intuitive but has its quirks—was genuinely painful at times.

Two bugs stand out. The first: the button kept returning a 404 and I was convinced the backend was at fault, but the request never reached it. The balancer only knew the `/` route, so it was answering `/random-image` itself. Worse, its response builder hardcoded `HTTP/1.1 200 OK`, so that "404" arrived at the browser as a *successful* response and the page cheerfully rendered the error text as an image path. The second: the server created `./images` but read from `../images`—two different folders, depending on where the process was launched from. Both were good reminders that when you write HTTP by hand, nothing is going to catch these mistakes for you.

I also simplified the image flow. It used to take two round trips: one to ask for a random image's path, another to fetch that path. That only works because every backend currently reads the same folder—once each instance has its own filesystem, the balancer could route the two requests to different backends and ask instance B for a file only instance A has. Now `/random-image` returns the image bytes directly, with the file name riding along in an `X-Image-Name` header for the page to display: one click, one request, one backend. I added a guard against path traversal as well, since a raw socket server joins user input straight onto a file path with nothing in between.

The proxy foundation is laid, but it currently forwards everything to a single hardcoded backend at `localhost:8081`, and answers with a 502 when that one is down because it has nowhere else to turn. Turning that constant into a proper pool—with round-robin selection and health checks—is the job for the coming days.
