import Cookies from 'universal-cookie';

import api from '../../api';
import { TEST, LOGIN_ACTION } from '../_constants';

const cookies = new Cookies();

/**
 * Create a dispatch object
 * @param {Object} payload
 */
function loginDispatch(payload) {
  return { type: LOGIN_ACTION, payload };
}

export const loginAction = (payload) => {
  if (TEST) return dispatch => dispatch(loginDispatch({ authenticated: true }));

  const { username, password } = payload;
  return async (dispatch) => {
    try {
      const token = await api.login(username, password);
      cookies.set('token', token);
      return dispatch(loginDispatch({ authenticated: true }));
    } catch (error) {
      cookies.remove('token');
      return dispatch(loginDispatch({ error }));
    }
  };
};
