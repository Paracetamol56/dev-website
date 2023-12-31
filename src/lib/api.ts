import axios from "axios";
import { user } from "./store";
import { get } from "svelte/store";
import { addToast } from "../routes/+layout.svelte";
import { goto } from "$app/navigation";
// import { isExpired } from "$lib/token";

async function login(email: string) {
  await axios.post(`/api/auth/login`, { email })
    .then((response) => {
      if (response.status != 200) {
        console.error(response);
        addToast({
          data: {
            title: 'Error',
            description: 'An error occured, please try again later',
            color: 'bg-ctp-red'
          }
        });
        return;
      }
      addToast({
        data: {
          title: 'Magic link sent',
          description: 'Please check your inbox',
          color: 'bg-ctp-green'
        }
      });
    })
    .catch((error) => {
      console.error(error);
      addToast({
        data: {
          title: 'Error',
          description: 'An error occured, please try again later',
          color: 'bg-ctp-red'
        }
      });
    });
}

async function verify(token: string, redirect: string) {
  await axios.post(`/api/auth/verify`, { token })
    .then((reponse) => {
      if (reponse.status !== 200) {
        console.error(reponse);
        addToast({
          data: {
            title: 'Invalid token',
            description: 'The token you provided is invalid or has expired, please try again',
            color: 'bg-ctp-red'
          }
        });
        goto(redirect);
        return;
      }
      // Write the user to the store
      user.set({
        accessToken: reponse.data.accessToken,
        refreshToken: reponse.data.refreshToken,
        id: reponse.data.user.id,
        email: reponse.data.user.email,
        flavour: reponse.data.user.flavour
      });

      addToast({
        data: {
          title: 'Successfully logged in',
          description: 'You are now logged in',
          color: 'bg-ctp-green'
        }
      });
      goto(redirect);
    })
    .catch(() => {
      addToast({
        data: {
          title: 'Failed to verify token',
          description: 'An error occured while verifying your token',
          color: 'bg-ctp-red'
        }
      });
      goto(redirect);
    });
}

function logout() {
  // Revoke tokens
  user.set({
    ...get(user),
    id: null,
    accessToken: null,
    refreshToken: null
  });
}

async function refreshAccessToken() {
  const refreshToken = get(user).refreshToken;
  if (!refreshToken) {
    addToast({
      data: {
        title: 'Outdated token',
        description: 'Your token is outdated, please log in again',
        color: 'bg-ctp-red'
      }
    });
    logout();
    return;
  }
  await axios.post(`/api/auth/refresh`, { refreshToken })
    .then(res => {
      if (res.status === 200) {
        // Update the user store
        user.set({
          ...get(user),
          accessToken: res.data.accessToken
        });
      }
    })
    .catch(err => {
      console.log(err);
      logout();
    });
}

async function callWithAuth(method: string, path: string, json: any) {
  const accessToken = get(user).accessToken;
  if (!accessToken) {
    await refreshAccessToken();
  } else {
    // isExpired(accessToken) && await refreshAccessToken();
  }
  return await axios({
    method,
    url: `/api${path}`,
    headers: {
      Authorization: `Bearer ${accessToken}`
    },
    data: json
  });
}

async function call(method: string, path: string, json: any) {
  return await axios({
    method,
    url: `/api${path}`,
    data: json
  });
}

export default { login, verify, logout, callWithAuth, call };
