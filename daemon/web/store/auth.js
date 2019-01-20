import Cookies from 'universal-cookie';

import {
  LOGIN_ACTION,
  LOGOUT_ACTION,
} from '../actions/_constants';

const cookies = new Cookies();

const Auth = (state = {
  authenticated: !!cookies.get('token'),
  expiry: null,
}, action) => {
  console.log('authReducer', { ...state, ...action.payload });
  switch (action.type) {
    case LOGIN_ACTION: return { ...state, ...action.payload };
    case LOGOUT_ACTION: return { ...state, ...action.payload };

    default: return state;
  }
};

export default Auth;
