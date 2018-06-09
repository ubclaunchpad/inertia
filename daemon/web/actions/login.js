import {
  TEST_LOGIN_ACTION,
} from './_constants';


export const testAction = payload => (dispatch) => {
  // remove later
  console.log('login action fired!');
  dispatch({
    type: TEST_LOGIN_ACTION,
    payload,
  });
};
