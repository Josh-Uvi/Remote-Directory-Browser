## Areas of focus

These are the areas we will be evaluating in the submission:

* Use consistent coding style. We use ESLint and format our code with
  [prettier](https://prettier.io/).
* Create a few unit-tests for scenarios you think make sense.
* Make sure builds are reproducible. Pick any vendoring/packaging system that
  will allow us to get consistent build results.
* Ensure error handling and error reporting is consistent. The app should report
  clear errors and not crash under non-critical conditions.
* Ensure that your app is secure.

The primary factor in the team's decision is overall code quality. We are looking for
the highest possible quality with the smallest possible scope that meets the requirements
of the challenge.

## Pitfalls and Gotchas

To help you out, we've composed a list of things that previously resulted in a
no-pass from the interview team:

* Use of AI. Don't outsource your thinking to an AI. We recommend using AI for
  use cases like learning about a new problem space, exploring APIs, and
  finding missing edge cases. However, we strongly recommend you write
  the design document and all code yourself.
* Scope creep. Candidates have tried to implement too much and ran out of time.
   * Avoid implementing an overly complex solution just to show that you are
     capable of writing a complex feature. Instead, if you think something could
     be made more complex in a full-fledged app, leave a comment about it and
     move on with a solution which solves the problem at hand.
   * For example, there is no need to implement a pluggable auth system which in
     the future would let you easily switch between different auth methods. It
     is better to focus on implementing a single auth method.
* Error handling. We pay extra attention to error handling. Make sure that they
  are properly handled and not ignored.
* Keep your CSS simple. We are not looking for animations or complex themes, but
  do want to see a responsive app that looks good on various screen sizes. Make
  sure that a long directory name doesn't break your layout.
* Make sure that your code is secured and your application is not vulnerable to
  common web security vulnerabilities.
    * Recommended resources:
        * [Authentication Cheat
          Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
        * [OWASP Top Ten](https://owasp.org/www-project-top-ten/)
* For a senior level, make sure you have a good crypto setup and secure session
  management.