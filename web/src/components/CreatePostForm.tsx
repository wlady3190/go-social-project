import React, { useState } from "react";
import { API_URL } from "../App";

const CreatePostForm: React.FC = () => {
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");

  const handleSubmit = async () => {
    await fetch(`${API_URL}/posts`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer foo`,
      },
      body: JSON.stringify({
        title,
        content,
      }),
    });
    setTitle("");
    setContent("");
  };
  return (
    <div>
      <div className="form">
        <label>
          <input
            type="text"
            placeholder="Title"
            onChange={(e) => setTitle(e.target.value)}
          />
        </label>
      </div>

      <label>
        <textarea
          name=""
          placeholder="Que piensas?"
          value={content}
          onChange={(e) => setContent(e.target.value)}
        ></textarea>
      </label>
      <button onClick={handleSubmit}>Send</button>
    </div>
  );
};

export default CreatePostForm;
