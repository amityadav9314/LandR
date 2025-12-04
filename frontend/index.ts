import { fetch } from 'expo/fetch';
import { registerRootComponent } from 'expo';

// Polyfill global fetch with Expo's implementation (supports streaming)
global.fetch = fetch as any;

import App from './App';

// registerRootComponent calls AppRegistry.registerComponent('main', () => App);
// It also ensures that whether you load the app in Expo Go or in a native build,
// the environment is set up appropriately
registerRootComponent(App);
