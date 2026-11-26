for (const form of document.querySelectorAll('form.guestbook-form')) {
    const submit = form.querySelector('button[type="submit"]');
    const feedback = form.querySelector('[data-guestbook-feedback]');
    const captchaBox = form.querySelector('[data-guestbook-captcha]');
    const captchaId = form.elements.namedItem('captcha_id');
    const captchaAnswer = form.elements.namedItem('captcha');
    const refresh = form.querySelector('[data-guestbook-refresh]');
    const recaptchaKey = form.dataset.recaptchaKey;

    function report(message, error = false) {
        feedback.textContent = message;
        feedback.toggleAttribute('data-error', error);
    }

    async function refreshCaptcha() {
        if (!refresh || refresh.disabled || captchaBox.hidden) return;
        refresh.disabled = true;
        captchaId.value = '';
        captchaAnswer.value = '';
        const image = refresh.querySelector('img');
        const label = refresh.querySelector('span');
        image.hidden = true;
        label.hidden = false;
        try {
            const response = await fetch('/api/captcha', { cache: 'no-store' });
            const result = await response.json();
            if (!response.ok || result.code !== 0) throw new Error();
            if (result.data.captcha_off) {
                captchaBox.hidden = true;
                captchaAnswer.required = false;
                return;
            }
            if (!result.data.captcha_id || !result.data.captcha) throw new Error();
            captchaId.value = result.data.captcha_id;
            image.src = result.data.captcha;
            image.hidden = false;
            label.hidden = true;
        } catch {
            report('验证码加载失败，请点击刷新验证码后重试。', true);
        } finally {
            refresh.disabled = false;
        }
    }

    async function getRecaptchaToken() {
        if (!window.grecaptcha) throw new Error();
        await new Promise((resolve) => window.grecaptcha.ready(resolve));
        return window.grecaptcha.execute(recaptchaKey, { action: 'guestbook' });
    }

    if (refresh) {
        refresh.addEventListener('click', refreshCaptcha);
        refreshCaptcha();
    }

    form.addEventListener('submit', async (event) => {
        event.preventDefault();
        if (submit.disabled) return;
        if (captchaAnswer?.required && !captchaId.value) {
            report('请先刷新并填写验证码。', true);
            return;
        }

        submit.disabled = true;
        report('正在提交…');
        let posted = false;
        try {
            const data = new FormData(form);
            data.set('return', 'json');
            if (recaptchaKey) data.set('recaptcha_token', await getRecaptchaToken());
            posted = true;
            const response = await fetch(form.action, { method: 'POST', body: data });
            const result = await response.json();
            if (!response.ok || result.code !== 0) {
                report(result.msg || '留言提交失败，请检查后重试。', true);
                return;
            }
            form.reset();
            report(result.msg || '留言已提交。');
        } catch {
            report(posted ? '未能确认提交结果，请稍后重试。' : '安全验证暂时不可用，请刷新页面后重试。', true);
        } finally {
            if (posted && refresh) await refreshCaptcha();
            submit.disabled = false;
        }
    });
}
