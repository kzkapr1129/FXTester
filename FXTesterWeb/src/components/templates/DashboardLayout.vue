<script setup lang="ts">
  import axios from 'axios';
import { ref } from 'vue';
  import { useRouter } from 'vue-router'
  const isOpenedDrawer = ref(false);
  const router = useRouter()

  const onClickNav = async (path: string) => {
    axios.delete("https://localhost:8000/api/auth")
    router.push(path)
  }

</script>

<template>
  <v-app>
    <v-system-bar color="secondary">
      System Bar
    </v-system-bar>

    <v-app-bar color="primary">
      <template v-slot:prepend>
          <v-app-bar-nav-icon
            @click.stop="isOpenedDrawer = !isOpenedDrawer">
          </v-app-bar-nav-icon>
      </template>

      <v-app-bar-title>
        Application Bar
      </v-app-bar-title>

      <template v-slot:append>
        <v-btn @click.stop="() => onClickNav('/login')">
          Sign-out
        </v-btn>

        <v-btn>
          <v-icon>
            mdi-home
          </v-icon>
        </v-btn>
      </template>
    </v-app-bar>

    <v-navigation-drawer app temporary v-model="isOpenedDrawer">
      <v-list>
        <v-list-item prepend-icon="mdi-view-dashboard" title="Home" value="home" @click.stop="() => onClickNav('/')"></v-list-item>
        <v-list-item prepend-icon="mdi-forum" title="About" value="about" @click.stop="() => onClickNav('/about')"></v-list-item>
      </v-list>
    </v-navigation-drawer>

    <v-main>
      <v-container>
        <RouterView />
      </v-container>
    </v-main>

    <v-bottom-navigation>
      Button Navigation
    </v-bottom-navigation>

    <v-footer color="primary" app>
      Footer
    </v-footer>

  </v-app>
</template>