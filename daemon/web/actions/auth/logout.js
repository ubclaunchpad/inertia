import { LOGOUT_ACTION } from '../_constants';

export const logoutAction = payload => (dispatch) => {
  dispatch({
    type: LOGOUT_ACTION,
    payload,
  });
};
