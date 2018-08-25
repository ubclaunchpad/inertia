import {
  LOGIN_ACTION,
  LOGOUT_ACTION,
} from '../actions/_constants';

const initialState = {
  authenticated: false,
  expiry: undefined,
};

const Login = (state = initialState, action) => {
  switch (action.type) {
    case LOGIN_ACTION: {
      return { ...state, authenticated: true };
    }
    case LOGOUT_ACTION: {
      return { ...state, authenticated: false };
    }

    default: {
      return state;
    }
  }
};

export default Login;
