# React Application (instaman)

This directory contains a small SPA that connects with go-instaman and provides a nice UI to automate and schedule operations on the main Instagram account.

The application is powered by:

* [Ant](https://ant.design)
* [React](https://react.dev)
* [TypeScript](https://www.typescriptlang.org)
* [Vite](https://vitejs.dev)

## Developer set up

The development version of this app runs on [node](https://nodejs.org) via [npm](https://www.npmjs.com).
Make sure both are installed, then spin it up with:

```sh
npm install
npm run dev
```

## App Directories

### `/public`

Static assets that are served straight from nginx. These are not included in the bundle.

### `/src/components`

All the React components are contained in this folder's subdirectories.

### `/src/api`

This directory contains the client, the types, and the interfaces required to interact with the go-instaman backend.
