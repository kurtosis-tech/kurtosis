# Enclave Manager UI (EM UI)

This codebase produces the enclave manager UI (ie `kurtosis web`). The `packages/web/src` directory contains:

- `client/enclaveManager` - libraries for interacting with the local `kurtosis` backend - used to instantiate a `KurtosisClientContext` and interacted with using `useKurtosisClient`
- `client/packageIndexer` - libraries for interacting with the package indexer - used to instantiate a `KurtosisPackageIndexerClientContext` and interacted with using `useKurtosisPackageIndexerClient`
- `emui` - the composition of the above to produce the Enclave Manager UI using react router. The code in here is generally structured into:
  - 'Routes' - used by react router to build the application
  - Whole page files (like EnclaveList.tsx) which structure each page. These are in a directory structure that is similar to the directory structure a user navigates when they use the enclave manager application.
  - 'components' - enclave manager specific components.

The other package in this repo, `packages/components` produces components used by this application, catalog.kurtosis.com and internal tooling.

## Available Scripts

In the project directory, you can run:

### `yarn cleanInstall`

Removes `node_modules` and runs `yarn install`.

### `yarn clean`

Removes the build output if present.

### `yarn start`

Runs the app in the development mode.\
Open [http://localhost:4000](http://localhost:4000) to view it in the browser.

The page will reload if you make edits.\
You will also see any lint errors in the console.

### `yarn test`

Launches the test runner in the interactive watch mode.\
See the section about [running tests](https://facebook.github.io/create-react-app/docs/running-tests) for more information.

### `yarn build`

Builds the app for production to the `build` folder.\
It correctly bundles React in production mode and optimizes the build for the best performance.

The build is minified and the filenames include the hashes.\
Your app is ready to be deployed!

See the section about [deployment](https://facebook.github.io/create-react-app/docs/deployment) for more information.

### `yarn start:prod`

Serve your local build on port 4000.

### `yarn prettier`

Runs `prettier --check` to check that the code matches the formatting that [`prettier`](https://prettier.io/) would apply.

### `yarn prettier:fix`

Applies any formatting changes prettier wants to apply to this application. For ease of use you can use IDE integrations
to auto apply prettier changes on file save, see instructions:

- [Here](https://plugins.jetbrains.com/plugin/10456-prettier) for Intellij
- [Here](https://marketplace.visualstudio.com/items?itemName=esbenp.prettier-vscode) for vscode

### `yarn cypress:ete`

Run the cypress ETE tests

### `yarn cypress:open`

Open the cypress console for debugging/developing new cypress tests. The cypress test suite runs against the locally
served EMUI from port 9711.

### `yarn eject`

**Note: this is a one-way operation. Once you `eject`, you can’t go back!**

If you aren’t satisfied with the build tool and configuration choices, you can `eject` at any time. This command will remove the single build dependency from your project.

Instead, it will copy all the configuration files and the transitive dependencies (webpack, Babel, ESLint, etc) right into your project so you have full control over them. All of the commands except `eject` will still work, but they will point to the copied scripts so you can tweak them. At this point you’re on your own.

You don’t have to ever use `eject`. The curated feature set is suitable for small and middle deployments, and you shouldn’t feel obligated to use this feature. However we understand that this tool wouldn’t be useful if you couldn’t customize it when you are ready for it.

## Development notes

### Key libraries used

- Typescript
- React
- react-scripts - these packages were created with Create React App.
- React router (v6) - we avoid using route actions and loaders as they didn't seem to have good Typescript support, created codepaths that were tricky to follow (by excessively fragmenting code) and generally seemed to increase complexity rather than improve development velocity.
- Chakra UI
- TanStack table - used for managing the state of tables throughout the application
- React hook form - used for forms where any validation is required. Generally the Smart Form Component pattern is followed to create form components that can be composed together arbitrarily.
- Reactflow - used for the graph in the enclave builder.
- Virtuoso - used for rendering streams of logs without dumping all of the text into the DOM.
- Monaco - used anywhere we want to display blocks of code, or for editing code in the browser.
- connect-es - this is used to create the client used for accessing kurtosis enclave manager and package indexer api's.

### Gotchas

- When you first start building this codebase you need to have a local build of the `enclave-manager-sdk`. This is because the `enclave-manager-sdk` is not a published library. The easiest way to make sure this is in place is to run `scripts/build.sh` from the root of this repository, or `enclave-manager/api/typescript/scripts/build.sh`.
- When you run cypress tests locally the tests are running against the app served by your local engine, not the one running from your development server. An easy way to temporarily work around this is to modify the port used from `9711` to `4000` in `cypress/supports/commands.ts`
- The wrapping library we use (`@monaco/react`) is a handy abstraction around Monaco, but using plugins with it seems very challenging without ejecting unfortunately.
- We use `react-mentions` for mention inputs in the enclave builder. This works fine, but better autocomplete could be built using Monaco.
- The enclave builder shows a graph of nested forms - because rendering it can be quite expensive, sensible use of `memo` around functional components in this part of the codebase can be particularly important. The enclave builder has two data models - one is the VariableContext which contains all of the form state across the graph, and provides the universe of available variables to mention in inputs; the other is the reactflow internal state which tracks node positions, types and sizes on the graph. The data models are linked by using the same `id` to refer to nodes.
- There is an issue with stream connections to the backend where they can be unexpectedly disconnected in some browsers, it's unclear what is causing this, but from [this bug](https://github.com/connectrpc/connect-es/issues/907) it seems like it could be in the kurtosis enclave manager server.
