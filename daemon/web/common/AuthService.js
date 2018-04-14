

const TOKEN_KEY = 'inertia';

export const login = () => {
  // set jwt
};

export const logout = () => {
  // clear jwt
};

export const isAuthenticated = () => {
  // verifies that jwt is still valid
  return true;
};

export const guardRoute = () => {
  if (!isAuthenticated()) {
    // push route
  }
};

export const setToken = (token) => {
  // store jwt in cookie
};

export const getToken = () => {
  // retrieve from cookie
};

export const removeToken = () => {
  // clear from cookie
};

export const isExpired = () => {
  // check if token is expired
};
