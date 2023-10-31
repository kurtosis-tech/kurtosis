# Enclave Manager UI (EMUI)

This codebase produces the enclave manager UI (ie `kurtosis web`). The `src` directory contains:

- `components` - components used in the application. This includes theme definitions and application context definitions
- `client/enclaveManager` - libraries for interacting with the local `kurtosis` backend - used to instantiate a `KurtosisClientContext` and interacted with using `useKurtosisClient`
- `client/packageIndexer` - libraries for interacting with the package indexer - used to instantiate a `KurtosisPackageIndexerClientContext` and interacted with using `useKurtosisPackageIndexerClient`
- `emui` - the composition of the above to produce the Enclave Manager UI using react router

## Available Scripts

In the project directory, you can run:

### `yarn cleanInstall`

Removes `node_modules` and runs `yarn install`.

### `yarn clean`

Removes the build output if present.

### `yarn start`

Runs the app in the development mode.\
Open [http://localhost:3000](http://localhost:3000) to view it in the browser.

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

### `yarn eject`

**Note: this is a one-way operation. Once you `eject`, you can’t go back!**

If you aren’t satisfied with the build tool and configuration choices, you can `eject` at any time. This command will remove the single build dependency from your project.

Instead, it will copy all the configuration files and the transitive dependencies (webpack, Babel, ESLint, etc) right into your project so you have full control over them. All of the commands except `eject` will still work, but they will point to the copied scripts so you can tweak them. At this point you’re on your own.

You don’t have to ever use `eject`. The curated feature set is suitable for small and middle deployments, and you shouldn’t feel obligated to use this feature. However we understand that this tool wouldn’t be useful if you couldn’t customize it when you are ready for it.
