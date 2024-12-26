import { useNavigate, useParams } from "react-router-dom";
import { API_URL } from "../App";

const ConfirmationPage = () => {
  const { token = "" } = useParams();
  const redirect = useNavigate();
  const handleConfirm = async () => {
    const response = await fetch(`${API_URL}/users/activate/${token}`, {
      method: "PUT",
    });

    if (response.ok) {
      //redirecto home page
      redirect("/");
    } else {
      //handle error
      alert("Failed to fecth");
    }
  };

  return (
    <>
      <h2>Confirmation</h2>
      <button onClick={handleConfirm}> CLick to confirm</button>
    </>
  );
};

export default ConfirmationPage;
