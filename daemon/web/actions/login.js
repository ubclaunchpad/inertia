import {
  TEST_LOGIN_ACTION,
} from './_constants';


export const testAction = payload => (dispatch) => {
  dispatch({
    type: TEST_LOGIN_ACTION,
    payload,
  });
};
