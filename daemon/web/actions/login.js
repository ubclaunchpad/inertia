import api from '../common/API';
import { TEST, LOGIN_ACTION } from './_constants';

const defaultDispatch = { type: LOGIN_ACTION };

export const loginAction = (payload) => {
  const { username, password } = payload;
  if (!username || !password) return dispatch => dispatch({
    ...defaultDispatch, payload: { authenticated: false },
  });
  if (TEST) return dispatch => dispatch({
    ...defaultDispatch, payload: { authenticated: true },
  });

  return async (dispatch) => {
    try {
      await api.login(username, password);
      return dispatch({ ...defaultDispatch, payload: { authenticated: true } });
    } catch (error) {
      return dispatch({ ...defaultDispatch, payload: { authenticated: false, error } });
    }
  };
};
