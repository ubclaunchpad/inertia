import {
  TEST_MAIN_ACTION,
} from './_constants';


export const testAction = payload => (dispatch) => {
  // remove later
  dispatch({
    type: TEST_MAIN_ACTION,
    payload,
  });
};
