module.exports = (api) => {
  const isProduction = api.env('production');
  return {
    presets: ['module:@react-native/babel-preset'],
    plugins: [
      // Strip all console.* calls in production builds to avoid leaking
      // sensitive data through logs and to reduce bundle overhead (L24).
      // Install the plugin: pnpm add -D babel-plugin-transform-remove-console
      ...(isProduction ? [['transform-remove-console', {exclude: ['error', 'warn']}]] : []),
    ],
  };
};
