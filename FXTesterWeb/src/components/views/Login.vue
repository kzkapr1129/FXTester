<script setup lang="ts">
import { onMounted } from 'vue';
import { useRoute} from 'vue-router'
import axios from 'axios'

const route = useRoute()
onMounted(() => {
  console.log("onMounted: ", window.location.origin, route.meta);
})
const onSignInClicked = async () => {
  const res = await axios({
    withCredentials: true,
    method: "POST",
    url: "https://127.0.0.1:8000/api/auth",
    data: {
      id: "admin",
      password: "admin"
    }
  });
  console.log("res hello:", res);
}
const onTestClicked = async () => {
  const res = await axios({
    withCredentials: true,
    method: "POST",
    url: "https://127.0.0.1:8000/api/auth/refresh",
    data: {}
  });
  console.log("res refresh:", res);
}
const onTest2Clicked = async () => {
  const res = await axios({
    withCredentials: true,
    method: "DELETE",
    url: "https://127.0.0.1:8000/api/auth",
    data: {}
  });
  console.log("res delete:", res);
}
</script>

<template>
  <v-app full-height>
    <v-main>
      <div class="d-flex justify-center align-center" :style="{height: '100%'}" id="login-page">
            <v-card min-width="440px" :style="{height: 'fit-content'}">
              <v-card-text>
                <div class="text-center mb-4">
                  <img src="../../assets/logo.png" width="80" height="80" />
                </div>

                <transition name="fade" mode="out-in">
                  <v-form ref="form">
                    <v-text-field
                      label="ユーザ名"
                      prepend-icon="mdi-account"
                      required
                      error-messages="ユーザ名が未入力です"
                    ></v-text-field>

                    <v-text-field
                      label="パスワード"
                      prepend-icon="mdi-lock"
                      type="password"
                      required
                    ></v-text-field>

                    <div class="text-center">
                      <v-btn
                        color="primary"
                        large
                        type="submit"
                        rounded
                        @click.stop.prevent="onSignInClicked"
                        >サインイン</v-btn>
                      <v-btn
                        color="primary"
                        large
                        type="submit"
                        rounded
                        @click.stop.prevent="onTestClicked"
                        >test</v-btn>
                      <v-btn
                        color="primary"
                        large
                        type="submit"
                        rounded
                        @click.stop.prevent="onTest2Clicked"
                        >ログアウト</v-btn>
                    </div>
                  </v-form>
                </transition>
              </v-card-text>
            </v-card>
      </div>
    </v-main>
  </v-app>
</template>

<style scoped>
#login-page {
  background-color: #34495e;
}
</style>