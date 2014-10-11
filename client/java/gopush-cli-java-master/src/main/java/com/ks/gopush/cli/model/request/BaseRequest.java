package com.ks.gopush.cli.model.request;

import java.io.Serializable;

/**
 *  基础请求
 * @author ghg
 *
 */
public class BaseRequest implements Serializable{

	private static final long serialVersionUID = 1L;
	/**
	 * 应用Token
	 */
	private String token;

	public String getToken() {
		return token;
	}

	public void setToken(String token) {
		this.token = token;
	}
	
}
