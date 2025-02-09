import Vue from 'vue'
import VueNativeSock from 'vue-native-websocket'
import App from './App.vue'
import store from './store'
import Howler from './plugins/Howler'

Vue.config.productionTip = false

let connectionString;
if (process.env.NODE_ENV === "production") {
  connectionString = `wss://${window.location.hostname}/ws`;
} else {
  connectionString = "ws://localhost:8081/ws";
}

Vue.use(VueNativeSock, connectionString, { store: store, format: 'json' })
Vue.use(Howler)

new Vue({
  store,
  render: h => h(App)
}).$mount('#app')
