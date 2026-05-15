## Requirements - Level 4

Implement an application that allows a user to browse directory content on a
remote server.

This application should have the following functionality:

* A Go backend that serves the webapp and an API
* The UI, which should include client-side filtering and sorting capabilities
  and URL-based navigation.
* Strong authentication

Additionally, we are a security-focused company and place an extra emphasis on
security for senior engineering candidates. We will expect your solution to have
a strong security posture as it pertains to authentication, encryption, and
overall web security.

### API

The repository we invite you to includes just enough starter code to serve up
both the web app and a sample API endpoint.

The starter code here only listens for plain-text HTTP. At Teleport, we avoid
plain-text connections and prefer to encrypt all data in transit. You will be
responsible for adding TLS.

Your API only needs to return the contents of the specified directory and does
not need to recurse into subdirectories. The following is an example of an
acceptable API response:

```json
{
  "name": "example",
  "type": "dir",
  "size": 0,

  "contents": [
    {
      "name": "README.md",
      "type": "file",
      "size": 12345,
    },
    {
      "name": "images",
      "type": "dir",
      "size": 0,
    }
  ]
}
```

When you add authentication (the requirements for which are outlined later in
this document), the API will also need to support session management (logging in
and out).

### UI

The UI should allow a user to view the contents of a single directory. Clicking
on a subdirectory should navigate to that directory and refresh the contents.
Unlike other commercial tools, _file preview is not required_. Clicking on a
file should not do anything.

The following features are required:

* [ ] Display the filename, type (file or directory), and human-readable size for files.
* [ ] Add support for filtering the directory contents based on filename.
  Filtering should be performed client-side, and a simple substring match is
  sufficient.
* [ ] Add support for sorting directory contents based on filename, type, and
  size.
* [ ] Include breadcrumbs that show the current location in the directory. The
  breadcrumbs should be clickable for easy navigation to parent directories.
* [ ] Implement URL navigation. The state of the app should be encoded in the
  URL. No state should be lost upon a page refresh.

While third-party dependencies are acceptable, we prefer to minimize the use of
dependencies where possible. For example:

* We prefer you use native browser APIs like
  [fetch](https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API) over
  third-party libraries like Axios.
* We're happy to see borrowed CSS from other projects to get a nice look and feel,
  but we do want the opportunity to evaluate how you design reusable components.
  Please implement your own components rather than using large frameworks that
  already provide components for breadcrumbs, tree-views, etc.

### Authentication

The app should also present directory information only to authenticated users.
This will require changes to both the API and the UI.

The following features are required:

* [ ] The API should reject requests from unauthenticated users
* [ ] The UI should redirect unauthenticated users to a login page. After a
  successful login, users should be directed back to the page they initially
  requested.
* [ ] The UI should provide a way for users to log out.

User sessions can be stored in memory, there is no need for a database of any
kind.

User registration / enrollment is not required. You are welcome to hard-code one
or two valid users, just let us know what their credentials are for testing
purposes.

Please do not use a third party solution that provides authentication out of
the box.