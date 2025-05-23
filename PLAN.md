# PLAN

End date: 2025-05-23

> [!NOTE]
> Write issues and milestones for all those when you can.
> 
> E.g. https://github.com/riraum/project451/milestone/1

- Milestone 00 - Due 2025-03-12 - Project setup
  - Find a name
  - Rename the repository
  - Setup CI
  - Write a `script/run` script that runs the code
  - Write a dummy `main.go`
  - Write a dummy `main_test.go`
  - Write in issues at least milestone 01 and 02

- Milestone 01 - Due 2025-03-19 - Basic website
  - Write a webserver that respond to `GET /` with `200 "OK"`
  - Write a router to handle responses to
    - `GET /` => `200 "OK"`
    - `GET /api/v0/posts` => `200 "[]"`
    - `POST /api/v0/posts` => `201`
    - Other paths should result in => `404`
  - Write tests for each route
  - Resources:
    - https://pkg.go.dev/net/http#hdr-Servers
    - https://pkg.go.dev/net/http#Server
    - https://pkg.go.dev/net/http#ServeMux
    - https://http.cat/

- Milestone 02 - Due 2025-03-26 - Basic DB
  - Create an SQLite db
  - Define a simple schema with a single table
    - `Post`: `id, date, title, link`
  - Write go code to initialize the database with the table
  - Write go code to write random data into the db  
  - Write go code to read all posts from the db
  - Write tests against a test db
  - Resources:
    - https://go.dev/doc/database/
    - https://pkg.go.dev/database/sql
    - https://go.dev/wiki/SQLDrivers
    - https://github.com/mattn/go-sqlite3 (most popular lib)
    - https://github.com/mattn/go-sqlite3/blob/master/_example/simple/simple.go

- Milestone 03 - Due 2025-04-02 - Basic frontend rendering
  - Update the `GET /` endpoint to:
    - List all the posts
    - Create and fill an html template
    - Create a basic form to submit new Posts (don't handle it yet)
  - Apply some basic CSS
  - Serve the CSS from your machine
  - Resources:
    - https://picocss.com/ (just a basic CSS framework that looks OK)
    - https://pkg.go.dev/html/template
    - https://www.alexedwards.net/blog/serving-static-sites-with-go
    - https://gist.github.com/paulmach/7271283

- Milestone 04 - Due 2025-04-09 - CRUD Posts
  - Update `POST /api/v0/posts` to accept the new data from the form and create it in DB
  - Add `DELETE /api/v0/posts/<id>` to delete a post from the DB
  - Update `GET /api/v0/posts` to accept the query parameter:
    - `sort`: `title`, `date` (default: `date`)
    - `direction`: `asc`, `desc` (default: `desc`)
    - E.g.
      - `GET /api/v0/posts?sort=title&direction=asc` list posts, order by title, Z->A
      - `GET /api/v0/posts?direction=asc` list posts, order by date, oldest -> newest

- Milestone 05 - Due 2025-04-16 - Posts Content and page

    - Update the post table to contain a field `content`
    - Add content form field when creating a post
    - Add a page to view a single post
    - Add a page to edit a single post

- Milestone 06 - Due 2025-04-23 - Authoring

    - Add a new table `Author: id, name`
    - Update the post table with `author` that contains the ID of the author
    - Create a couple of authors that the app knows by default
    - Create a couple of posts per authors
    - Add a search query `author=X` that filter only for this author, by default
      it search for all authors.

- Milestone 07 - Due 2025-04-30 - Authentication

    - Setup a login page that contains a single field: the author's name.
      When the author's name is correctly sent via `POST`, authenticate the
      current user. This is done by setting a cookie.
    - For all requests, get the cookie value, and ensure that it's a valid
      autor's name, if not, you should fail with 401.
    - Allow to create a post only when authenticated, and set the author during
      the post submission on the backend.

    - Refs:
      - https://pkg.go.dev/net/http#Request.Cookie
      - https://pkg.go.dev/net/http#SetCookie

- Milestone 08 - Due 2025-05-07 - TBD

- Milestone 09 - Due 2025-04-14 - TBD

- Milestone 10 - Due 2025-04-21 - TBD

- Milestone 11 - Due 2025-04-25 - Final review and cleanup
