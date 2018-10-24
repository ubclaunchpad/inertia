import { VALIDATE_ACTION } from '../_constants';

export const validateAction = payload => (dispatch) => {
  dispatch({
    type: VALIDATE_ACTION,
    payload,
  });
};
