# Contributing

## Setup your machine

The OpenAList backend is written in [Go](https://go.dev/), and the frontend is built with [SolidJS](https://www.solidjs.com/).

Prerequisites:

- [git](https://git-scm.com)
- [Go 1.25+](https://go.dev/doc/install)
- [gcc](https://gcc.gnu.org/)
- [Node.js 22.22.1+](https://nodejs.org/)
- [pnpm 11](https://pnpm.io/installation)

Clone the backend and frontend as sibling directories:

```shell
git clone https://github.com/AlliotTech/openalist.git
git clone --recurse-submodules https://github.com/AlliotTech/openalist-web.git
```
You should switch to the `main` branch for development.

Install the frontend dependencies:

```shell
cd openalist-web
pnpm install --frozen-lockfile
```

## Preview your change

Start the backend and frontend together from the backend repository:

```shell
cd ../openalist
./dev.sh
```

The script expects the frontend at `../openalist-web`. Set `OPENALIST_WEB_DIR` if it is stored elsewhere, and set `OPENALIST_DATA_DIR` to use a different development data directory.

To start only the backend:

```shell
go run . server --dev --data ./data-dev
```

To start only the frontend:

```shell
cd ../openalist-web
pnpm dev
```

Run the backend tests before submitting a pull request:

```shell
go test ./...
```

## Add a new driver
Copy `drivers/template` folder and rename it, and follow the comments in it.

## Create a commit

Commit messages should be well formatted, and to make that "standardized".

### Commit Message Format
Each commit message consists of a **header**, a **body** and a **footer**.  The header has a special
format that includes a **type**, a **scope** and a **subject**:

```
<type>(<scope>): <subject>
<BLANK LINE>
<body>
<BLANK LINE>
<footer>
```

The **header** is mandatory and the **scope** of the header is optional.

Any line of the commit message cannot be longer than 100 characters! This allows the message to be easier
to read on GitHub as well as in various git tools.

### Revert
If the commit reverts a previous commit, it should begin with `revert: `, followed by the header
of the reverted commit.
In the body it should say: `This reverts commit <hash>.`, where the hash is the SHA of the commit
being reverted.

### Type
Must be one of the following:

* **feat**: A new feature
* **fix**: A bug fix
* **docs**: Documentation only changes
* **style**: Changes that do not affect the meaning of the code (white-space, formatting, missing
  semi-colons, etc)
* **refactor**: A code change that neither fixes a bug nor adds a feature
* **perf**: A code change that improves performance
* **test**: Adding missing or correcting existing tests
* **build**: Affects project builds or dependency modifications
* **revert**: Restore the previous commit
* **ci**: Continuous integration of related file modifications
* **chore**: Changes to the build process or auxiliary tools and libraries such as documentation
  generation
* **release**: Release a new version

### Scope
The scope could be anything specifying place of the commit change. For example `$location`,
`$browser`, `$compile`, `$rootScope`, `ngHref`, `ngClick`, `ngView`, etc...

You can use `*` when the change affects more than a single scope.

### Subject
The subject contains succinct description of the change:

* use the imperative, present tense: "change" not "changed" nor "changes"
* don't capitalize first letter
* no dot (.) at the end

### Body
Just as in the **subject**, use the imperative, present tense: "change" not "changed" nor "changes".
The body should include the motivation for the change and contrast this with previous behavior.

### Footer
The footer should contain any information about **Breaking Changes** and is also the place to
[reference GitHub issues that this commit closes](https://help.github.com/articles/closing-issues-via-commit-messages/).

**Breaking Changes** should start with the word `BREAKING CHANGE:` with a space or two newlines.
The rest of the commit message is then used for this.

## Submit a pull request

Push your branch to your `alist` fork and open a pull request against the
`main` branch.
