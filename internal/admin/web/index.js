const UI = {
  getModal: (elementId) =>
    bootstrap.Modal.getInstance(document.getElementById(elementId)),
  setButtonState: (btn, disabled, text) => {
    btn.disabled = disabled;
    if (text) btn.textContent = text;
  },
  showError: (errorDiv, input, message) => {
    errorDiv.textContent = message;
    errorDiv.style.display = "block";
    input?.classList.add("is-invalid");
  },
  clearError: (errorDiv, input) => {
    errorDiv.style.display = "none";
    errorDiv.textContent = "";
    input?.classList.remove("is-invalid");
  },
  resetInput: (input, errorDiv) => {
    input.value = "";
    if (errorDiv) UI.clearError(errorDiv, input);
  },
};

async function safeJsonParse(response) {
  try {
    return await response.json();
  } catch (error) {
    console.error("Failed to parse JSON response:", error);
    return { error: "Invalid server response" };
  }
}

async function apiCall(url, options = {}) {
  try {
    const response = await fetch(url, options);

    if (response.status === 401) {
      window.location.href = "/admin/login";
      return null;
    }

    if (!response.ok && response.status >= 500) {
      console.error(`HTTP ${response.status} error for ${url}:`, {
        status: response.status,
        statusText: response.statusText,
        url,
      });
    }

    return response;
  } catch (error) {
    console.error(`Network error for ${url}:`, error);
    throw new Error(
      "Unable to connect to the server. Please check your connection.",
    );
  }
}

function showToast(message, type = "success") {
  const toast = document.createElement("div");
  toast.className = `notification-toast ${type}`;
  toast.textContent = message;
  document.body.appendChild(toast);
  setTimeout(() => {
    toast.classList.add("fade-out");
    setTimeout(() => toast.remove(), 300);
  }, 3000);
}

function showAlert(title, message) {
  const modal = document.createElement("div");
  modal.className = "modal-alert show";
  modal.innerHTML = `
    <div class="modal-alert-content">
      <div class="modal-alert-title">${title}</div>
      <div class="modal-alert-message">${message}</div>
      <div class="modal-alert-buttons">
        <button class="modal-alert-btn-primary">OK</button>
      </div>
    </div>
  `;
  modal.querySelector(".modal-alert-btn-primary").onclick = () =>
    modal.remove();
  modal.addEventListener("click", (e) => e.target === modal && modal.remove());
  document.body.appendChild(modal);
}

function showConfirm(title, message, onConfirm, onCancel) {
  const modal = document.createElement("div");
  modal.className = "modal-alert show";
  modal.innerHTML = `
    <div class="modal-alert-content">
      <div class="modal-alert-title">${title}</div>
      <div class="modal-alert-message">${message}</div>
      <div class="modal-alert-buttons">
        <button class="modal-alert-btn-secondary">Cancel</button>
        <button class="modal-alert-btn-danger">Confirm</button>
      </div>
    </div>
  `;
  const buttons = modal.querySelectorAll("button");
  buttons[0].onclick = () => {
    modal.remove();
    onCancel?.();
  };
  buttons[1].onclick = () => {
    modal.remove();
    onConfirm?.();
  };
  modal.addEventListener("click", (e) => {
    if (e.target === modal) {
      modal.remove();
      onCancel?.();
    }
  });
  document.body.appendChild(modal);
}

let currentPage = null;

function loadPage(page) {
  currentPage = page;

  const navLinks = document.querySelectorAll(".nav-link");
  navLinks.forEach((link) => link.classList.remove("active"));

  const activeLink = document.querySelector(`a[onclick="loadPage('${page}')"]`);
  if (activeLink) {
    activeLink.classList.add("active");
  }

  fetch(`/admin/${page}`)
    .then((r) => r.text())
    .then((html) => {
      document.getElementById("content").innerHTML = html;
      if (page === "tokens") {
        loadTokens();
      } else if (page === "devices") {
        loadDevices();
      }
    })
    .catch((e) => {
      console.error("Error loading page:", e);
      document.getElementById("content").innerHTML =
        '<div class="content-main"><div class="alert alert-danger">Unable to load page. Please try refreshing.</div></div>';
    });
}

function formatDate(dateString) {
  if (!dateString) return "-";
  try {
    const date = new Date(dateString);
    if (isNaN(date.getTime())) return dateString;
    return date.toLocaleDateString() + " " + date.toLocaleTimeString();
  } catch {
    return dateString;
  }
}

function copyToClipboard(text) {
  navigator.clipboard
    .writeText(text)
    .then(() => showToast("Copied to clipboard"))
    .catch(() => showAlert("Error", "Failed to copy to clipboard"));
}

function copyFromElement(elementId) {
  const elem = document.getElementById(elementId);
  if (elem) copyToClipboard(elem.value || elem.textContent);
}

function maskString(str) {
  if (!str) return "";
  if (str.length <= 4) return "****";
  const visible = str.slice(-4);
  return "*".repeat(str.length - 4) + visible;
}

function toggleField(elementId, fieldType = "field") {
  const elem = document.getElementById(elementId);
  if (!elem) return;
  const btn = document.querySelector(
    `button[onclick="toggleField('${elementId}')"]`,
  );

  if (elem.dataset.revealed === "false") {
    elem.textContent = elem.dataset.value;
    elem.dataset.revealed = "true";
    if (btn) btn.innerHTML = "👁️‍🗨️";
  } else {
    elem.textContent = maskString(elem.dataset.value);
    elem.dataset.revealed = "false";
    if (btn) btn.innerHTML = "👁️";
  }
}

function showCreateTokenModal() {
  document.getElementById("createTokenForm").reset();
  document.getElementById("useHost").checked = false;
  document.getElementById("setExpiry").checked = false;
  document.getElementById("expiryDate").style.display = "none";
  document.getElementById("expiryDate").disabled = true;
  document.getElementById("attachToSession").checked = false;
  UI.setButtonState(
    document.getElementById("createTokenBtn"),
    false,
    "Create Token",
  );
  new bootstrap.Modal(document.getElementById("createTokenModal")).show();
}

function toggleExpiryDateInput() {
  const expiryInput = document.getElementById("expiryDate");
  if (document.getElementById("setExpiry").checked) {
    expiryInput.style.display = "block";
    expiryInput.disabled = false;
    const sevenDaysLater = new Date();
    sevenDaysLater.setDate(sevenDaysLater.getDate() + 7);
    expiryInput.value = sevenDaysLater.toISOString().slice(0, 16);
  } else {
    expiryInput.style.display = "none";
    expiryInput.disabled = true;
    expiryInput.value = "";
  }
}

async function loadTokens() {
  try {
    const response = await apiCall("/api/v1/admin/tokens");
    if (!response) return;
    const tokens = await safeJsonParse(response);
    const tbody = document.getElementById("tokensBody");
    if (!tbody) return;

    if (!tokens?.length) {
      tbody.innerHTML =
        '<tr><td colspan="7" class="table-empty-state">No tokens found</td></tr>';
      return;
    }

    tbody.innerHTML = tokens
      .map(
        (token) => `
      <tr>
        <td>
          <div style="display: flex; align-items: center; gap: 5px;">
            <code id="username-${token.id}" data-value="${token.matrix_username}" data-revealed="false" style="font-size: 0.9em;">${maskString(token.matrix_username)}</code>
            <button class="btn btn-sm btn-outline-secondary p-0" onclick="toggleField('username-${token.id}')" title="Toggle visibility" style="width: 24px; height: 24px; line-height: 16px; border: 1px solid var(--color-border);">👁️</button>
            <button class="btn btn-sm btn-outline-secondary p-0" onclick="copyToClipboard('${token.matrix_username}')" title="Copy" style="width: 24px; height: 24px; line-height: 16px; border: 1px solid var(--color-border);">📋</button>
          </div>
        </td>
        <td>
          <div style="display: flex; align-items: center; gap: 5px;">
            <code id="device-${token.id}" data-value="${token.matrix_device_id}" data-revealed="false" style="font-size: 0.9em;">${maskString(token.matrix_device_id)}</code>
            <button class="btn btn-sm btn-outline-secondary p-0" onclick="toggleField('device-${token.id}')" title="Toggle visibility" style="width: 24px; height: 24px; line-height: 16px; border: 1px solid var(--color-border);">👁️</button>
            <button class="btn btn-sm btn-outline-secondary p-0" onclick="copyToClipboard('${token.matrix_device_id}')" title="Copy" style="width: 24px; height: 24px; line-height: 16px; border: 1px solid var(--color-border);">📋</button>
          </div>
        </td>
        <td>${token.is_admin ? '<span class="badge bg-warning">Yes</span>' : '<span class="badge bg-secondary">No</span>'}</td>
        <td>${token.expires_at ? formatDate(token.expires_at) : "-"}</td>
        <td>${token.last_used_at ? formatDate(token.last_used_at) : "-"}</td>
        <td>${formatDate(token.created_at)}</td>
        <td><button class="btn btn-sm btn-outline-danger" onclick="deleteToken(${token.id})">Delete</button></td>
      </tr>
    `,
      )
      .join("");
  } catch (error) {
    console.error("Error loading tokens:", error);
    const tbody = document.getElementById("tokensBody");
    if (tbody)
      tbody.innerHTML =
        '<tr><td colspan="7" class="table-empty-state text-danger">Unable to load tokens. Please try refreshing the page.</td></tr>';
  }
}

async function createToken() {
  try {
    const btn = document.getElementById("createTokenBtn");
    const originalText = btn.textContent;
    UI.setButtonState(btn, true, "Creating...");

    const body = { use_host: document.getElementById("useHost").checked };

    if (document.getElementById("setExpiry").checked) {
      const expiryDateInput = document.getElementById("expiryDate").value;
      if (!expiryDateInput) {
        showAlert("Validation Error", "Please select an expiry date");
        UI.setButtonState(btn, false, originalText);
        return;
      }
      body.expires_at = new Date(expiryDateInput).toISOString();
    }

    const response = await apiCall("/api/v1/admin/tokens", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    if (!response) return;

    if (response.ok) {
      const data = await safeJsonParse(response);
      const attachToSession =
        document.getElementById("attachToSession").checked;
      UI.getModal("createTokenModal").hide();
      document.getElementById("createTokenForm").reset();
      document.getElementById("displayToken").value = data.token;
      document.getElementById("attachToSessionCheckbox").checked =
        attachToSession;
      document.getElementById("attachToSessionSection").style.display =
        attachToSession ? "block" : "none";
      new bootstrap.Modal(document.getElementById("tokenDisplayModal")).show();
      document
        .getElementById("tokenDisplayModal")
        .addEventListener("hidden.bs.modal", loadTokens, { once: true });
    } else {
      const error = await safeJsonParse(response);
      console.error("Failed to create token:", {
        status: response.status,
        error,
      });
      showAlert(
        "Error",
        error.error || "Unable to create token. Please try again later.",
      );
      UI.setButtonState(btn, false, originalText);
    }
  } catch (error) {
    console.error("Error creating token:", error);
    showAlert(
      "Error",
      "Unable to create token. Please check your connection and try again.",
    );
    UI.setButtonState(
      document.getElementById("createTokenBtn"),
      false,
      "Create Token",
    );
  }
}

function closeTokenDisplayModal() {
  const shouldAttach = document.getElementById(
    "attachToSessionCheckbox",
  )?.checked;
  if (shouldAttach) {
    const token = document.getElementById("displayToken").value.trim();
    setMatrixTokenFromDisplay(token);
  }
  UI.getModal("tokenDisplayModal").hide();
}

async function setMatrixTokenFromDisplay(token) {
  try {
    if (!token) {
      showAlert("Error", "No token to attach. Please create a token first.");
      return;
    }
    const response = await apiCall("/api/v1/admin/matrix-token", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ token }),
    });
    if (!response) return;
    if (response.ok) {
      showToast("Token attached to session", "success");
    } else {
      const error = await safeJsonParse(response);
      console.error("Failed to attach token:", {
        status: response.status,
        error,
      });
      showAlert(
        "Error",
        error.error || "Failed to attach token. Please try again.",
      );
    }
  } catch (error) {
    console.error("Error attaching token:", error);
    showAlert(
      "Error",
      "Unable to attach token. Please check your connection and try again.",
    );
  }
}

function toggleExpiryDateInput() {
  const expiryInput = document.getElementById("expiryDate");
  if (document.getElementById("setExpiry").checked) {
    expiryInput.style.display = "block";
    expiryInput.disabled = false;
    const sevenDaysLater = new Date();
    sevenDaysLater.setDate(sevenDaysLater.getDate() + 7);
    expiryInput.value = sevenDaysLater.toISOString().slice(0, 16);
  } else {
    expiryInput.style.display = "none";
    expiryInput.disabled = true;
    expiryInput.value = "";
  }
}

async function deleteToken(id) {
  showConfirm(
    "Delete Token",
    "Are you sure you want to delete this token? This action cannot be undone.",
    async () => {
      try {
        const response = await apiCall(`/api/v1/admin/tokens/${id}`, {
          method: "DELETE",
        });
        if (!response) return;
        if (response.ok) {
          loadTokens();
          showToast("Token deleted successfully", "success");
        } else {
          const error = await safeJsonParse(response);
          console.error("Failed to delete token:", {
            status: response.status,
            error,
          });
          showAlert(
            "Error",
            error.error || "Unable to delete token. Please try again.",
          );
        }
      } catch (error) {
        console.error("Error deleting token:", error);
        showAlert(
          "Error",
          "Unable to delete token. Please check your connection and try again.",
        );
      }
    },
  );
}

async function loadDevices() {
  try {
    const hasTokenResponse = await apiCall("/api/v1/admin/matrix-token-status");
    if (!hasTokenResponse) return;
    const hasTokenData = await safeJsonParse(hasTokenResponse);

    if (!hasTokenData.has_matrix_token) {
      const modal = new bootstrap.Modal(
        document.getElementById("setMatrixTokenModal"),
      );
      modal.show();
      document.getElementById("devicesList").innerHTML =
        '<div class="alert alert-info">Please set your Matrix token to continue.</div>';
      return;
    }

    const response = await apiCall("/api/v1/admin/devices");
    if (!response) return;

    const devices = await safeJsonParse(response);
    const devicesList = document.getElementById("devicesList");

    if (!devicesList) return;

    if (!devices || devices.length === 0) {
      devicesList.innerHTML =
        '<table class="table"><tbody><tr><td class="table-empty-state" colspan="3">No devices found</td></tr></tbody></table>';
      return;
    }

    const deviceRows = devices
      .map(
        (d, idx) => `<tr>
<td><span class="platform-badge">${d.platform}</span></td>
<td>
  <div style="display: flex; align-items: center; gap: 5px;">
    <code id="device-${idx}" data-value="${d.device_id}" data-revealed="false" style="font-size: 0.9em;">${maskString(d.device_id)}</code>
    <button class="btn btn-sm btn-outline-secondary p-0" onclick="toggleField('device-${idx}')" title="Toggle visibility" style="width: 24px; height: 24px; line-height: 16px;">👁️</button>
    <button class="btn btn-sm btn-outline-secondary p-0" onclick="copyToClipboard('${d.device_id}')" title="Copy to clipboard" style="width: 24px; height: 24px; line-height: 16px;">📋</button>
  </div>
</td>
<td>
  <div style="display: flex; gap: 6px; align-items: center;">
    <button class="btn btn-sm btn-outline-primary" onclick="showSendMessageModal('${d.device_id}', '${d.platform}')" title="Send message">
      Send Message
    </button>
    <button class="btn btn-sm btn-outline-danger" onclick="deleteDevice('${d.device_id}', '${d.platform}')" title="Delete device">
      Delete
    </button>
  </div>
</td>
</tr>`,
      )
      .join("");

    devicesList.innerHTML =
      '<table class="table">' +
      "<thead><tr><th>Platform</th><th>Device ID</th><th>Actions</th></tr></thead>" +
      `<tbody>${deviceRows}</tbody></table>`;
  } catch (error) {
    console.error("Error loading devices:", error);
    const devicesList = document.getElementById("devicesList");
    if (devicesList) {
      devicesList.innerHTML =
        '<table class="table"><tbody><tr><td class="table-empty-state text-danger" colspan="3">Unable to load devices. Please try refreshing the page.</td></tr></tbody></table>';
    }
  }
}

async function setMatrixToken() {
  try {
    const input = document.getElementById("matrixTokenInput");
    const errorDiv = document.getElementById("matrixTokenError");
    const btn = document.getElementById("setMatrixTokenBtn");
    const token = input.value.trim();

    if (!token) {
      UI.showError(errorDiv, input, "Matrix token is required");
      return;
    }

    if (!token.startsWith("mt_")) {
      UI.showError(errorDiv, input, "Token must start with mt_");
      return;
    }

    UI.clearError(errorDiv, input);
    UI.setButtonState(btn, true, "Setting...");

    const response = await apiCall("/api/v1/admin/matrix-token", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ token }),
    });
    if (!response) return;

    if (response.ok) {
      UI.getModal("setMatrixTokenModal").hide();
      UI.resetInput(input, errorDiv);
      showToast("Matrix token set successfully", "success");
      loadDevices();
    } else {
      const error = await safeJsonParse(response);
      console.error("Failed to set Matrix token:", {
        status: response.status,
        error,
      });
      UI.showError(errorDiv, input, error.error || "Failed to set token");
      UI.setButtonState(btn, false, "Set Token");
    }
  } catch (error) {
    console.error("Error setting Matrix token:", error);
    const errorDiv = document.getElementById("matrixTokenError");
    const input = document.getElementById("matrixTokenInput");
    UI.showError(
      errorDiv,
      input,
      "Unable to connect. Please check your connection and try again.",
    );
    UI.setButtonState(
      document.getElementById("setMatrixTokenBtn"),
      false,
      "Set Token",
    );
  }
}

function skipMatrixTokenModal() {
  const input = document.getElementById("matrixTokenInput");
  const errorDiv = document.getElementById("matrixTokenError");
  UI.resetInput(input, errorDiv);
  UI.setButtonState(
    document.getElementById("setMatrixTokenBtn"),
    false,
    "Set Token",
  );
  UI.getModal("setMatrixTokenModal").hide();
}

async function deleteDevice(deviceId, platform) {
  showConfirm(
    "Delete Device",
    "Are you sure you want to delete this device?",
    async () => {
      try {
        const response = await apiCall("/api/v1/admin/devices", {
          method: "DELETE",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ device_id: deviceId, platform }),
        });
        if (!response) return;

        if (response.ok) {
          loadDevices();
          showToast("Device deleted successfully", "success");
        } else {
          const error = await safeJsonParse(response);
          console.error("Failed to delete device:", {
            status: response.status,
            error,
          });
          showAlert(
            "Error",
            error.error || "Unable to delete device. Please try again.",
          );
        }
      } catch (error) {
        console.error("Error deleting device:", error);
        showAlert(
          "Error",
          "Unable to delete device. Please check your connection and try again.",
        );
      }
    },
  );
}

function showAddDeviceModal() {
  new bootstrap.Modal(document.getElementById("addDeviceModal")).show();
  resetAddDeviceModal();
}

function resetAddDeviceModal() {
  document.getElementById("platformSelectionStep").style.display = "block";
  document.getElementById("qrCodeStep").style.display = "none";
  document.getElementById("connectionStatus").style.display = "flex";
  document.getElementById("successMessage").style.display = "none";
  document.getElementById("qrCanvas").innerHTML =
    '<div class="qr-loading-spinner"></div>';

  const platformBtn = document.getElementById("platformBtn");
  const loadingIndicator = document.getElementById("platformLoadingIndicator");
  if (platformBtn) platformBtn.disabled = false;
  if (loadingIndicator) loadingIndicator.style.display = "none";

  if (window.currentWebSocket) window.currentWebSocket.close();
  window.currentWebSocket = null;
  window.currentQRCodeUrl = null;
}

async function selectPlatform(platform) {
  window.selectedPlatform = platform;

  const platformBtn = document.getElementById("platformBtn");
  const loadingIndicator = document.getElementById("platformLoadingIndicator");

  platformBtn.disabled = true;
  loadingIndicator.style.display = "block";

  try {
    const response = await apiCall("/api/v1/admin/devices", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ platform }),
    });

    if (!response) return;

    if (!response.ok) {
      const error = await safeJsonParse(response);
      console.error("Failed to add device:", {
        status: response.status,
        error,
      });
      showAlert(
        "Error",
        error.error || "Unable to add device. Please try again.",
      );
      platformBtn.disabled = false;
      loadingIndicator.style.display = "none";
      return;
    }

    const data = await safeJsonParse(response);
    document.getElementById("platformSelectionStep").style.display = "none";
    document.getElementById("qrCanvas").innerHTML =
      '<div class="qr-loading-spinner"></div>';
    document.getElementById("qrCodeStep").style.display = "block";

    const qrCodeUrl = data.qr_code_url;
    if (
      !qrCodeUrl ||
      (qrCodeUrl.includes("token=") && qrCodeUrl.endsWith("token="))
    ) {
      showAlert(
        "Error",
        "Failed to generate QR code: empty token. Please ensure your Matrix token is set.",
      );
      resetAddDeviceModal();
      return;
    }

    await new Promise((resolve) => setTimeout(resolve, 5000));
    connectWebSocket(qrCodeUrl);
  } catch (error) {
    console.error("Error creating device:", error);
    showAlert(
      "Error",
      "Unable to add device. Please check your connection and try again.",
    );
    platformBtn.disabled = false;
    loadingIndicator.style.display = "none";
    resetAddDeviceModal();
  }
}

function retryQRConnection() {
  if (window.currentQRCodeUrl) {
    document.getElementById("qrCanvas").innerHTML =
      '<div class="qr-loading-spinner"></div>';
    document.getElementById("statusText").textContent = "Reconnecting...";
    if (window.currentWebSocket) {
      window.currentWebSocket.close();
      window.currentWebSocket = null;
    }
    connectWebSocket(window.currentQRCodeUrl);
  }
}

function connectWebSocket(qrCodeUrl) {
  try {
    window.currentQRCodeUrl = qrCodeUrl;
    if (window.currentWebSocket) window.currentWebSocket.close();

    window.currentWebSocket = new WebSocket(qrCodeUrl);
    window.currentWebSocket.receivedData = false;
    window.currentWebSocket.hasError = false;

    window.currentWebSocket.onopen = () => {
      document.getElementById("statusText").textContent =
        "Waiting for device...";
    };

    window.currentWebSocket.onmessage = (event) => {
      window.currentWebSocket.receivedData = true;
      if (event.data.startsWith("Error:")) {
        window.currentWebSocket.hasError = true;
        document.getElementById("statusText").innerHTML =
          `${event.data} <button class="btn-retry" onclick="retryQRConnection()">Try again</button>`;
        document.getElementById("qrCanvas").innerHTML =
          '<div class="qr-retry-box" onclick="retryQRConnection()" style="cursor: pointer;"><div class="retry-icon">↻</div><div class="retry-text">Retry</div></div>';
      } else {
        generateQRCode(event.data);
      }
    };

    window.currentWebSocket.onerror = () => {
      document.getElementById("statusText").innerHTML =
        'Connection error. <button class="btn-retry" onclick="retryQRConnection()">Try again</button>';
      document.getElementById("qrCanvas").innerHTML =
        '<div class="qr-retry-box" onclick="retryQRConnection()" style="cursor: pointer;"><div class="retry-icon">↻</div><div class="retry-text">Retry</div></div>';
    };

    window.currentWebSocket.onclose = () => {
      if (
        window.currentWebSocket &&
        window.currentWebSocket.receivedData &&
        !window.currentWebSocket.hasError
      )
        showDeviceConnected();
    };
  } catch (error) {
    console.error("Error connecting to WebSocket:", error);
    document.getElementById("statusText").innerHTML =
      'Connection error. <button class="btn-retry" onclick="retryQRConnection()">Try again</button>';
  }
}

function generateQRCode(data) {
  const container = document.getElementById("qrCanvas");
  container.innerHTML = "";

  new QRCode(container, {
    text: data,
    width: 300,
    height: 300,
    colorDark: "#000000",
    colorLight: "#ffffff",
    correctLevel: QRCode.CorrectLevel.H,
  });
}

function showDeviceConnected() {
  document.getElementById("connectionStatus").style.display = "none";
  document.getElementById("successMessage").style.display = "flex";
  setTimeout(() => {
    UI.getModal("addDeviceModal")?.hide();
    loadDevices();
    showToast("Device added successfully!", "success");
  }, 2000);
}

function showSendMessageModal(deviceId, platform) {
  window.selectedDeviceId = deviceId;
  window.selectedDevicePlatform = platform;
  document.getElementById("messageContact").value = "";
  document.getElementById("messageText").value = "";
  UI.setButtonState(document.getElementById("sendMessageBtn"), false, "Send");
  new bootstrap.Modal(document.getElementById("sendMessageModal")).show();
}

async function sendMessage() {
  const contact = document.getElementById("messageContact").value.trim();
  const text = document.getElementById("messageText").value.trim();

  if (!contact) {
    showAlert("Validation Error", "Contact number is required");
    return;
  }
  if (!text) {
    showAlert("Validation Error", "Message is required");
    return;
  }

  const btn = document.getElementById("sendMessageBtn");
  try {
    UI.setButtonState(btn, true, "Sending...");
    const response = await apiCall(
      `/api/v1/admin/devices/${window.selectedDeviceId}/message`,
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          contact,
          platform: window.selectedDevicePlatform,
          text,
        }),
      },
    );

    if (!response) return;

    if (response.ok) {
      UI.getModal("sendMessageModal").hide();
      showToast("Message queued successfully", "success");
    } else {
      const error = await safeJsonParse(response);
      console.error("Failed to send message:", {
        status: response.status,
        error,
      });
      showAlert(
        "Error",
        error.error || "Unable to send message. Please try again.",
      );
      UI.setButtonState(btn, false, "Send");
    }
  } catch (error) {
    console.error("Error sending message:", error);
    showAlert(
      "Error",
      "Unable to send message. Please check your connection and try again.",
    );
    UI.setButtonState(btn, false, "Send");
  }
}

document.addEventListener("DOMContentLoaded", () => {
  loadPage("tokens");
});
